package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// Message is the standard WebSocket broadcast message.
type Message struct {
	Type    string `json:"type"`
	GroupID int64  `json:"groupId"`
	MatchID int64  `json:"matchId,omitempty"`
	Payload any    `json:"payload,omitempty"`
}

// Client represents a connected WebSocket client subscribed to an event.
type Client struct {
	EventID int64
	conn    *websocket.Conn
	send    chan []byte
}

// Hub maintains active WebSocket clients grouped by eventID.
type Hub struct {
	// eventID → set of clients
	rooms      map[int64]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastReq
}

type broadcastReq struct {
	eventID int64
	data    []byte
}

// NewHub creates and returns a new Hub.
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]bool),
		register:   make(chan *Client, 32),
		unregister: make(chan *Client, 32),
		broadcast:  make(chan broadcastReq, 256),
	}
}

// Run is the main event loop for the Hub — must be called in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			if h.rooms[c.EventID] == nil {
				h.rooms[c.EventID] = make(map[*Client]bool)
			}
			h.rooms[c.EventID][c] = true

		case c := <-h.unregister:
			if room, ok := h.rooms[c.EventID]; ok {
				if room[c] {
					delete(room, c)
					close(c.send)
				}
				if len(room) == 0 {
					delete(h.rooms, c.EventID)
				}
			}

		case req := <-h.broadcast:
			room, ok := h.rooms[req.eventID]
			if !ok {
				continue
			}
			for c := range room {
				select {
				case c.send <- req.data:
				default:
					// Slow client — drop and unregister.
					delete(room, c)
					close(c.send)
				}
			}
		}
	}
}

// BroadcastToEvent serialises msg and sends it to all clients in the event room.
func (h *Hub) BroadcastToEvent(eventID int64, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("hub.BroadcastToEvent marshal: %v", err)
		return
	}
	h.broadcast <- broadcastReq{eventID: eventID, data: data}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// NewClient creates a new WebSocket client and starts its write pump.
func NewClient(eventID int64, conn *websocket.Conn) *Client {
	return &Client{
		EventID: eventID,
		conn:    conn,
		send:    make(chan []byte, 256),
	}
}

// WritePump pumps messages from the send channel to the WebSocket connection.
func (c *Client) WritePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// ReadPump keeps the connection alive by reading (and discarding) pong/control frames.
func (c *Client) ReadPump(unregister func(*Client)) {
	defer func() {
		unregister(c)
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
