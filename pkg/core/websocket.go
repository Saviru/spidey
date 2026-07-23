package core

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512 KB
)



type SpideyConn struct {
	ID        string
	Conn      *websocket.Conn
	Hub       *WSHub
	Send      chan []byte
	onMessage func(msg []byte)
}

// registers a callback for incoming WebSocket messages
func (c *SpideyConn) OnMessage(handler func(msg []byte)) {
	c.onMessage = handler
}

// manages active WebSocket connections and broadcasting (with optional Redis Pub/Sub)
type WSHub struct {
	clients    map[string]map[*SpideyConn]bool // Channel/Room -> Clients
	register   chan *subscription
	unregister chan *subscription
	broadcast  chan *message
	mu         sync.RWMutex

	rdb      *redis.Client
	redisCtx context.Context
}

type subscription struct {
	conn    *SpideyConn
	channel string
}

type message struct {
	data    []byte
	channel string
}

// creates a new WebSocket Hub
func NewWSHub() *WSHub {
	h := &WSHub{
		clients:    make(map[string]map[*SpideyConn]bool),
		register:   make(chan *subscription),
		unregister: make(chan *subscription),
		broadcast:  make(chan *message),
		redisCtx:   context.Background(),
	}
	go h.run()
	return h
}

// enables Redis Pub/Sub for horizontal scaling across multiple servers
func (h *WSHub) UseRedis(addr, password string, db int) {
	h.rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	log.Println("Spidey WSHub: Connected to Redis for Pub/Sub")
}

func (h *WSHub) run() {
	for {
		select {
		case sub := <-h.register:
			h.mu.Lock()
			if h.clients[sub.channel] == nil {
				h.clients[sub.channel] = make(map[*SpideyConn]bool)
				// If Redis is enabled, subscribe to the channel
				if h.rdb != nil {
					go h.redisSubscribe(sub.channel)
				}
			}
			h.clients[sub.channel][sub.conn] = true
			h.mu.Unlock()

		case sub := <-h.unregister:
			h.mu.Lock()
			if connections, ok := h.clients[sub.channel]; ok {
				if _, ok := connections[sub.conn]; ok {
					delete(connections, sub.conn)
					close(sub.conn.Send)
					if len(connections) == 0 {
						delete(h.clients, sub.channel)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			// If Redis is enabled, publish to Redis INSTEAD of broadcasting locally
			// The redisSubscribe goroutine will receive it and broadcast it locally
			if h.rdb != nil {
				h.rdb.Publish(h.redisCtx, "spidey:ws:"+msg.channel, msg.data)
			} else {
				h.localBroadcast(msg.channel, msg.data)
			}
		}
	}
}

func (h *WSHub) localBroadcast(channel string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[channel] {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clients[channel], client)
		}
	}
}

func (h *WSHub) redisSubscribe(channel string) {
	pubsub := h.rdb.Subscribe(h.redisCtx, "spidey:ws:"+channel)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for msg := range ch {
		h.localBroadcast(channel, []byte(msg.Payload))
	}
}

// sends a message to a specific channel
func (h *WSHub) BroadcastTo(channel string, data []byte) {
	h.broadcast <- &message{data: data, channel: channel}
}

// pumps messages from the websocket connection to the hub.
func (c *SpideyConn) readPump() {
	defer func() {
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Spidey WebSocket error: %v", err)
			}
			break
		}
		if c.onMessage != nil {
			c.onMessage(msg)
		}
	}
}

// pumps messages from the hub to the websocket connection.
func (c *SpideyConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// adds a connection to a specific channel
func (c *SpideyConn) Subscribe(channel string) {
	c.Hub.register <- &subscription{conn: c, channel: channel}
}

// removes a connection from a specific channel
func (c *SpideyConn) Unsubscribe(channel string) {
	c.Hub.unregister <- &subscription{conn: c, channel: channel}
}
