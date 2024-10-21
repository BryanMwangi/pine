// Pine's websocket package is a websocket server that supports multiple channels
// This feature is experimental and may change in the future.
// Please use it with caution and at your own risk.

package websocket

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/gorilla/websocket"
)

// Config is a struct that holds the configuration for the websocket server
type Config struct {
	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. If a buffer
	// size is zero, then buffers allocated by the HTTP server are used. The
	// I/O buffer sizes do not limit the size of the messages that can be sent
	// or received.
	ReadBufferSize, WriteBufferSize int

	// Subprotocols specifies the server's supported protocols in order of
	// preference. If this field is not nil, then the Upgrade method negotiates a
	// subprotocol by selecting the first match in this list with a protocol
	// requested by the client. If there's no match, then no protocol is
	// negotiated (the Sec-Websocket-Protocol header is not included in the
	// handshake response).
	SubprotocolsAllowed []string

	// CheckOrigin returns true if the request Origin header is acceptable. If
	// CheckOrigin is nil, then a safe default is used: return false if the
	// Origin request header is present and the origin host is not equal to
	// request Host header.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	CheckOrigin func(r *http.Request) bool

	// Error specifies the function for generating HTTP error responses. If Error
	// is nil, then http.Error is used to generate the HTTP response.
	Error func(w http.ResponseWriter, r *http.Request, status int, reason error)

	// EnableCompression specify if the server should attempt to negotiate per
	// message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool

	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout time.Duration

	// This defines the the type of connection you wish to create
	// it can be "self" or "managed"
	// if you set it to "self" you will need to use the New function to open a
	// new connection
	// if you set it to "managed" you will need to use the Managed function to open a
	// new connection
	//
	// default is "self"
	// Please not that "managed" is experimental and may change in the future
	Type string
}

var defaultConfig = Config{
	SubprotocolsAllowed: []string{""},
	EnableCompression:   true,
	HandshakeTimeout:    10 * time.Second,
	CheckOrigin:         func(r *http.Request) bool { return true },
	Error:               func(w http.ResponseWriter, r *http.Request, status int, reason error) {},
	ReadBufferSize:      4096,
	WriteBufferSize:     4096,
	Type:                "self",
}

// Conn is a struct that holds the websocket connection
type Conn struct {
	*websocket.Conn
	viewedBytesSize int
}

var poolConn = sync.Pool{
	New: func() interface{} {
		return new(Conn)
	},
}

// Acquire Conn from pool
func acquireConn() *Conn {
	conn := poolConn.Get().(*Conn)
	return conn
}

// Return Conn to pool
func releaseConn(conn *Conn) {
	conn.Conn = nil
	poolConn.Put(conn)
}

// Called to open a new connection and upgrade it to a websocket connection
// this is the main function to use to create a new websocket connection
// Use this function if the Type is set to "self"
func New(handler func(conn *Conn, ctx *pine.Ctx), config ...Config) pine.Handler {
	var cfg Config
	if len(config) > 0 {
		userConfig := config[0]
		if userConfig.ReadBufferSize != 0 {
			cfg.ReadBufferSize = userConfig.ReadBufferSize
		}
		if userConfig.WriteBufferSize != 0 {
			cfg.WriteBufferSize = userConfig.WriteBufferSize
		}
		if userConfig.SubprotocolsAllowed != nil {
			cfg.SubprotocolsAllowed = userConfig.SubprotocolsAllowed
		}
		if userConfig.CheckOrigin != nil {
			cfg.CheckOrigin = userConfig.CheckOrigin
		}
		if userConfig.Error != nil {
			cfg.Error = userConfig.Error
		}
		if userConfig.EnableCompression {
			cfg.EnableCompression = userConfig.EnableCompression
		}
		if userConfig.HandshakeTimeout != 0 {
			cfg.HandshakeTimeout = userConfig.HandshakeTimeout
		}
		if userConfig.Type != "" {
			cfg.Type = userConfig.Type
		}
	} else {
		cfg = defaultConfig
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		CheckOrigin:       cfg.CheckOrigin,
		Error:             cfg.Error,
		Subprotocols:      cfg.SubprotocolsAllowed,
		EnableCompression: cfg.EnableCompression,
		HandshakeTimeout:  cfg.HandshakeTimeout,
	}

	return func(ctx *pine.Ctx) error {
		Conn, err := upgrader.Upgrade(ctx.Response.ResponseWriter, ctx.Request, ctx.Response.Header())
		if err != nil {
			fmt.Println(err)
			return err
		}

		if cfg.Type != "self" {
			panic("ChannelType must be 'self'")
		}
		conn := acquireConn()
		conn.Conn = Conn
		defer releaseConn(conn)
		handler(conn, ctx)
		return nil
	}

}
