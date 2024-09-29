// Pine's websocket package is a websocket server that supports multiple channels
// This feature is experimental and may change in the future.
// Please use it with caution and at your own risk.
package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var RunTimeTree = ConnectionTree{
	Channels: make(map[uuid.UUID]*Channel),
	Clients:  make(map[uuid.UUID]Client),
}

type ConnectionTree struct {
	Channels map[uuid.UUID]*Channel
	Clients  map[uuid.UUID]Client
	CM       sync.RWMutex
}

type Client struct {
	Conn    *websocket.Conn
	Channel *Channel
	Id      uuid.UUID
	IP      string
	Send    chan []byte
}

type Channel struct {
	ID      uuid.UUID
	Clients []*Client
	Message chan []byte
	CM      sync.Mutex
}

var (
	ReconnectTimeout = 5 * time.Second
	WriteWait        = 10 * time.Second
	PongWait         = 60 * time.Second
	PingPeriod       = (PongWait * 9) / 10
	MaxRetryAttempts = 5
)

// This function is used to create a new channel and client
// It is called when a new connection is made by the managed function
func create(conn *websocket.Conn, ctx *pine.Ctx) *Channel {
	var client Client
	var c Channel

	ip := ctx.IP()
	// we check if the client is already registered
	for _, clt := range RunTimeTree.Clients {
		if clt.IP == ip {
			client = clt
			break
		}
	}

	if client.IP == "" {
		client = registerIP(ip, conn)
	}

	c.ID = uuid.New()
	c.Message = make(chan []byte, 100)
	c.Clients = append(c.Clients, &client)

	RunTimeTree.CM.Lock()
	defer RunTimeTree.CM.Unlock()
	RunTimeTree.Channels[c.ID] = &c

	client.Channel = &c

	go c.Broadcast()

	go client.readPump()
	go client.writePump()
	return &c
}

// used to register the client and associate their IP address to a UUID
// this is called when a new connection is made by the managed function
func registerIP(ip string, conn *websocket.Conn) Client {
	// we check if the client is already registered
	for _, client := range RunTimeTree.Clients {
		if client.IP == ip {
			// we return the client if it is already registered
			return client
		}
	}
	// we create a new client and associate it with the IP address
	client := Client{
		Id:   uuid.New(),
		IP:   ip,
		Conn: conn,
		Send: make(chan []byte, 100),
	}
	RunTimeTree.Clients[client.Id] = client
	return client
}

// used to remove a client from the connection tree
// avoid using this to manually remove clients
// use MoveClientToChannel instead
func (c *ConnectionTree) RemoveClient(clientID uuid.UUID) {
	delete(c.Clients, clientID)
}

// used to remove a client from a channel
// avoid using this to manually remove clients
//
// when a client disconnects, it is automatically removed from the channel
// no need to call this function manually
func (c *Channel) RemoveClientFromChannel(clientId uuid.UUID) {
	c.CM.Lock()
	defer c.CM.Unlock()
	for i, cl := range c.Clients {
		if cl.Id == clientId {
			c.Clients = append(c.Clients[:i], c.Clients[i+1:]...)
			break
		}
	}
	if len(c.Clients) == 0 {
		RunTimeTree.CM.Lock()
		defer RunTimeTree.CM.Unlock()
		delete(RunTimeTree.Channels, c.ID)
	}
}

// used to move a client to a new channel
// Use this function to move a client to a new channel especially when you want to
// when you want to manually move a client to a new channel
//
// Example: You want to move a client to a new channel when a user joins a chat room
func (c *Channel) MoveClientToChannel(client *Client, newChannel *Channel) error {
	newChannel.CM.Lock() // Lock channel before modification
	defer newChannel.CM.Unlock()
	_, ok := RunTimeTree.Channels[newChannel.ID]
	if ok {
		for i, cl := range c.Clients {
			if cl.Id == client.Id {
				c.Clients = append(c.Clients[:i], c.Clients[i+1:]...)
				break
			}
		}
		// Optionally remove channel if no clients remain
		if len(c.Clients) == 0 {
			RunTimeTree.CM.Lock()
			defer RunTimeTree.CM.Unlock()
			delete(RunTimeTree.Channels, c.ID)
		}
	}
	// only add unique clients to the new channel
	for _, cl := range newChannel.Clients {
		if client.Id == cl.Id || client.IP == cl.IP {
			// Client is already in the new channel, no need to add
			return nil
		}
	}

	// Add the client to the new channel
	newChannel.Clients = append(newChannel.Clients, client)
	client.Channel = newChannel // Update the client's channel reference

	return nil
}

// used to broadcast a message to all clients in the channel
// avoid calling this function manually as it is called automatically during the
// managed function runtime
func (c *Channel) Broadcast() {
	for message := range c.Message {
		// we check if there are any clients in the channel
		if len(c.Clients) == 0 {
			continue
		}
		for _, client := range c.Clients {
			client.Send <- message
		}
	}
}

// used to read the incoming messages from the client
// this is an internal function

func (c *Client) readPump() {
	defer func() {
		RunTimeTree.RemoveClient(c.Id)
		c.Channel.RemoveClientFromChannel(c.Id)
		c.Conn.Close()
	}()
	c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break // Exit loop on error
		}
		// Send the message to the channel's broadcast mechanism
		c.Channel.Message <- message
	}
}

// used to write the outgoing messages to the client
// this is an internal function
func (c *Client) writePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				// The channel was closed, so we send a close message
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
