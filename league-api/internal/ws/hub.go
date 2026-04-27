package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout   = 10 * time.Second
	historyMaxAge  = time.Hour
	historyMaxSize = 200
)

// Message is the standard WebSocket broadcast message.
type Message struct {
	Type      string    `json:"type"`
	GroupID   int64     `json:"groupId"`
	MatchID   int64     `json:"matchId,omitempty"`
	Payload   any       `json:"payload,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// historyEntry holds a serialised message and its timestamp for replay.
type historyEntry struct {
	timestamp time.Time
	data      []byte
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

	// per-eventID message history — protected by historyMu
	historyMu sync.RWMutex
	history   map[int64][]historyEntry
}

type broadcastReq struct {
	eventID   int64
	data      []byte
	timestamp time.Time
}

// NewHub creates and returns a new Hub.
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[int64]map[*Client]bool),
		register:   make(chan *Client, 32),
		unregister: make(chan *Client, 32),
		broadcast:  make(chan broadcastReq, 256),
		history:    make(map[int64][]historyEntry),
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
			// Append to history before broadcasting so a reconnecting client
			// calling MessagesSince concurrently sees it.
			h.appendHistory(req.eventID, req.timestamp, req.data)

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

// appendHistory adds an entry to the per-event ring buffer and prunes stale entries.
func (h *Hub) appendHistory(eventID int64, ts time.Time, data []byte) {
	h.historyMu.Lock()
	defer h.historyMu.Unlock()

	buf := h.history[eventID]

	// Trim entries older than historyMaxAge.
	cutoff := time.Now().UTC().Add(-historyMaxAge)
	start := 0
	for start < len(buf) && buf[start].timestamp.Before(cutoff) {
		start++
	}
	buf = buf[start:]

	// Enforce max size (ring semantics — drop oldest).
	if len(buf) >= historyMaxSize {
		buf = buf[len(buf)-historyMaxSize+1:]
	}

	buf = append(buf, historyEntry{timestamp: ts, data: data})
	h.history[eventID] = buf
}

// MessagesSince returns all buffered raw JSON messages for eventID with
// timestamp strictly after `since`, in chronological order.
func (h *Hub) MessagesSince(eventID int64, since time.Time) [][]byte {
	h.historyMu.RLock()
	defer h.historyMu.RUnlock()

	buf := h.history[eventID]
	if len(buf) == 0 {
		return nil
	}

	var out [][]byte
	for _, e := range buf {
		if e.timestamp.After(since) {
			out = append(out, e.data)
		}
	}
	return out
}

// BroadcastToEvent serialises msg (stamping Timestamp if zero) and sends it to
// all clients in the event room.
func (h *Hub) BroadcastToEvent(eventID int64, msg Message) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("hub.BroadcastToEvent marshal: %v", err)
		return
	}
	h.broadcast <- broadcastReq{eventID: eventID, data: data, timestamp: msg.Timestamp}
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
// A write deadline is set on every write to prevent slow/hung clients from
// blocking the goroutine indefinitely.
func (c *Client) WritePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			return
		}
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
