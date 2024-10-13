package pine

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Client is a wrapper around the http.Client
// Has methods for setting the request URI, method, headers, and body
//
// Ideally you want to use a client only once, however if you need to use
// multiple instances of the same client, start by creating a new client and
// then call the SetRequestURI, SetMethod, SetHeaders, and SetBody methods
//
// Note that if you call SendRequest, the old response will be overwritten
// If you would like to store the response, call ReadResponse and store the
// response in a variable of your own choosing
type Client struct {
	*http.Client
	req *Request
	res *http.Response
}

type Request struct {
	http.Request
	body        *bytes.Buffer
	uri         string
	method      string
	jsonEncoder JSONMarshal
}

// Common errors if you want to use the client and its methods
var (
	ErrURIRequired    = errors.New("uri is required")
	ErrMethodRequired = errors.New("method is required")
	ErrResponseIsNil  = errors.New("response is nil")
)

// Call this to create a new client
// You can then call SetRequestURI, SetMethod, SetHeaders, and SetBody
// after creating the client
func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
		req: &Request{
			jsonEncoder: json.Marshal,
		},
	}
}

// NewClientWithTimeout returns a new client with a timeout
func NewClientWithTimeout(timeout time.Duration) *Client {
	return &Client{
		Client: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Request() *Request {
	return c.req
}

// Sets the body of the request as JSON
func (r *Request) JSON(body interface{}) error {
	raw, err := r.jsonEncoder(body)
	if err != nil {
		return err
	}
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	// set the content type
	r.Header.Set("Content-Type", "application/json")

	// allows bytes to be streamed by the client similar to the io.Reader
	r.body = bytes.NewBuffer(raw)
	return nil
}

// Use this to set the headers of the request
// You can add as many headers as you want in a map
//
// For example:
//
//	headers := map[string]string{
//		"X-API-KEY": "1234567890",
//	}
//
// request.SetHeaders(headers)
func (r *Request) SetHeaders(headers map[string]string) {
	if r.Header == nil {
		r.Header = make(http.Header)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
}

// Use this to set the url of the request
//
// For example:
// request.SetRequestURI("https://example.com/api/v1/users")
func (r *Request) SetRequestURI(uri string) *Request {
	r.uri = uri
	return r
}

// Use this to set the method of the request
//
// For example:
// request.SetMethod("GET")
func (r *Request) SetMethod(method string) *Request {
	r.method = method
	return r
}

// Use this method to skip TLS verification
// This can be useful if the api you are calling has outdated TLS certificates
func (c *Client) SetTLSVerification(skip bool) {
	c.Client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skip},
	}
}

// Call this method only if you have already set the request URI and method
//
// body and headers are optional
//
// It returns an error if the request URI or method is not set
// Please note that if you call SendRequest, the old response will be overwritten
func (c *Client) SendRequest() error {
	if c.req.uri == "" {
		return ErrURIRequired
	}

	if c.req.method == "" {
		return ErrMethodRequired
	}

	var req *http.Request
	var err error

	if c.req.body == nil {
		req, err = http.NewRequest(c.req.method, c.req.uri, nil)
	} else {
		req, err = http.NewRequest(c.req.method, c.req.uri, c.req.body)
	}

	if err != nil {
		return err
	}
	for k, v := range c.req.Header {
		req.Header[k] = v
	}
	c.req.Request = *req

	res, err := c.Client.Do(&c.req.Request)
	if err != nil {
		return err
	}
	c.res = res
	return nil
}

// Call this method to get the response from the request
// Note that after calling this method the old reponse will be discarded
//
// Attempts to read the response after calling this method more than
// once will return an error
func (c *Client) ReadResponse() (code int, body []byte, err error) {
	if c.res == nil {
		return 0, nil, ErrResponseIsNil
	}
	resCode := c.res.StatusCode
	respBody := c.res.Body
	defer c.res.Body.Close()
	// we read the bytes from the res body stream
	buff := new(bytes.Buffer)
	_, err = buff.ReadFrom(respBody)

	// extract the body from the buffer
	body = buff.Bytes()
	// release the response after reading
	c.releaseResponse()
	return resCode, body, err
}

// Internal method used to release the response after reading it
func (c *Client) releaseResponse() {
	c.res = nil
}
