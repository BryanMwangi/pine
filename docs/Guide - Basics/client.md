---
sidebar_position: 4
---

# Client

Client is a wrapper around the http.Client. This implementation is no special and you can instead use the http.Client directly if you wish.

There are a couple of helpful methods that allow you to reduce the amount of code you need to write, so there's that.

Here is how the client struct looks like:

```go
type Client struct {
	*http.Client
	req *Request
	res *http.Response
}
```

The request and response are pretty standard. The Request is a simple wrapper around the http.Request struct with some extra properties, while the response is a direct wrapper around the http.Response struct.

```go
type Request struct {
    http.Request
    body        *bytes.Buffer
    uri         string
    method      string
    jsonEncoder JSONMarshal
}
```

By default, we use the json.Marshal for encoding the request body, however, if the need arises, let me know if you would like to use the same encoder as set in the [`Server config`](./pine#config).

## Request Methods

The following methods are available on the Request struct.

### SetMethod

This is used to set http method that will be used to send the request. Also supports chaining.

```go
func (c *Request) SetMethod(method string) *Request
```

### SetRequestURI

This is used to set the URI of the request. You will need to include the full uri and protocol to use this method. Also supports chaining.

```go
func (c *Request) SetRequestURI(uri string) *Request
```

Example:

```go
request.SetRequestURI("https://example.com/api/v1/users")
```

### SetHeaders

This method is used to set the headers of a request. You will need to construct a map of the headers and pass that map to the method. The map needs to be in the following type:

```go
map[string]string
```

Also supports chaining.

```go
func (c *Request) SetHeaders(headers map[string]string) *Request
```

### JSON

This method is used to set the body of the request to be JSON. You can pass any data here to be used by the JSON encoder. This method does not support chaining and returns an error.

It is preffered to call this method last to check for errors during encoding the JSON.

```go
func (c *Request) JSON(body interface{}) error
```

## Client Methods

The following methods are available on the Client struct.

### New

Not a method but a function that returns a new instance of the Client struct.

```go
func NewClient() *Client
```

### NewClientWithTimeout

Also not a method but a function that returns a new instance of the Client struct with a timeout.

```go
func NewClientWithTimeout(timeout time.Duration) *Client
```

### Request

This is a method and returns an instance of the Request struct and supports chaining, meaning you can then chain more request methods from this instance.

```go
func (c *Client) Request() *Request
```

### SetTLSVerification

You can use this method to skip TLS verification if the request is made over HTTPS. This is particularly useful when say you call an API that has outdated certificates.

```go
func (c *Client) SetTLSVerification(skip bool) *Client
```

### SendRequest

This method is used to send the request and returns an error if the request fails. Only use this method after you have set all the properties of the request.

The minimum required properties are the method and uri. The body is optional.

```go
func (c *Client) SendRequest() error
```

### ReadResponse

After a request is sent, by default the response is stored in the same instance of the request and can be retrieved and freed by calling ReadResponse.

After ReadResponse is called once, the response is no longer available to that specific instance of the request.

There was no specific reason for decoupling the response from the request method during my designing phase, however, you may consider that not each request warrants you to read the response, for example, you might just want to send a `ping` to show that the server is up and running.

Improvements to this method are welcome.

```go
func (c *Client) ReadResponse() error
```
