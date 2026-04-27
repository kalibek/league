package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"league-api/internal/model"
	"league-api/internal/repository"
	"league-api/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development; tighten in production.
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketHandler struct {
	hub       *ws.Hub
	eventRepo repository.EventRepository
}

func NewWebSocketHandler(hub *ws.Hub, eventRepo repository.EventRepository) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, eventRepo: eventRepo}
}

// GET /ws/events/:eid?since=<RFC3339|unix-ms>
func (h *WebSocketHandler) Handle(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	event, err := h.eventRepo.GetByID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	if event.Status != model.EventInProgress {
		c.JSON(http.StatusForbidden, gin.H{"error": "event is not in progress"})
		return
	}

	// Parse optional ?since= for catch-up replay.
	var since time.Time
	if sinceStr := c.Query("since"); sinceStr != "" {
		// Try RFC3339 first, then Unix milliseconds.
		if t, err := time.Parse(time.RFC3339Nano, sinceStr); err == nil {
			since = t
		} else if ms, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
			since = time.UnixMilli(ms).UTC()
		}
		// If parsing fails entirely we just ignore and replay nothing.
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // upgrader already wrote the error response
	}

	client := ws.NewClient(eventID, conn)
	h.hub.Register(client)

	// Replay missed messages before starting the normal pumps.
	if !since.IsZero() {
		missed := h.hub.MessagesSince(eventID, since)
		for _, msg := range missed {
			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				conn.Close()
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				conn.Close()
				return
			}
		}
	}

	// Start write pump in goroutine.
	go client.WritePump()

	// Read pump blocks until client disconnects.
	client.ReadPump(h.hub.Unregister)
}
