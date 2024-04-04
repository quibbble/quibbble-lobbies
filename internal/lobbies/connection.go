package lobbies

import (
	"context"
	"net"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	// playerMessageBuffer controls the max number
	// of messages that can be queued for a player
	// before it is kicked.
	playerMessageBuffer = 16

	// playerWriteTimeout determines how long to
	// wait before removing the player's connection.
	playerWriteTimeout = time.Second * 3
)

// Connection represents a player connected to the game server.
// Messages are sent on the messages channel and if the client
// cannot keep up with the messages, closeSlow is called.
type Connection struct {

	// player information
	player *Player

	// outputCh provides a channel the game server use to
	// send messages to the player.
	outputCh chan []byte

	// inputCh provides a channel the player can use to
	// send messages to the game server.
	inputCh chan *Action

	// conn is the underlying websocket connection between
	// the player and the game server.
	conn *websocket.Conn

	// closed represents whether or not the websocket
	// connection has been closed.
	closed bool

	// mu ensures closed is thread safe.
	mu sync.Mutex
}

func NewConnection(player *Player, conn *websocket.Conn, inputCh chan *Action) *Connection {
	return &Connection{
		player:   player,
		outputCh: make(chan []byte, playerMessageBuffer),
		inputCh:  inputCh,
		conn:     conn,
	}
}

func (c *Connection) ReadPump(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return net.ErrClosed
	}
	c.mu.Unlock()

	for {
		var msg Message
		if err := wsjson.Read(ctx, c.conn, &msg); err != nil {
			return err
		}
		c.inputCh <- &Action{
			Message:    &msg,
			Connection: c,
		}
	}
}

func (c *Connection) WritePump(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return net.ErrClosed
	}
	c.mu.Unlock()

	for {
		select {
		case msg := <-c.outputCh:
			if err := c.write(ctx, msg); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.conn != nil {
		c.conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
	}
}

func (c *Connection) write(ctx context.Context, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, playerWriteTimeout)
	defer cancel()
	return c.conn.Write(ctx, websocket.MessageText, msg)
}
