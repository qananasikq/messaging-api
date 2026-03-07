// client.go

package websocket

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	closed atomic.Bool
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, 128),
	}
}

func (c *Client) TrySend(b []byte) {
	if c.closed.Load() {
		return
	}
	select {
	case c.send <- b:
	default:
		log.Println("Failed to send message to client, buffer full")
	}
}

func (c *Client) WritePump(ctx context.Context) {
	defer c.Close(CloseReasonServerShutdown)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.conn.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				log.Println("Error sending message:", err)
				return
			}
		}
	}
}

func (c *Client) ReadPump(ctx context.Context) {
	defer c.Close(CloseReasonServerShutdown)

	for {
		rctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		_, _, err := c.conn.Read(rctx)
		cancel()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}
	}
}

func (c *Client) Close(reason CloseReason) {
	if !c.closed.Load() {
		c.conn.Close(websocket.StatusNormalClosure, string(reason))
		c.closed.Store(true)
	}
}
