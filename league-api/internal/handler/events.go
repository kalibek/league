package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"league-api/internal/model"
	"league-api/internal/service"
)

type EventsHandler struct {
	eventSvc   service.EventService
	draftSvc   service.DraftService
	leagueSvc  service.LeagueService
	groupRepo  interface {
		ListByEvent(ctx interface{}, eventID int64) ([]model.Group, error)
		GetPlayers(ctx interface{}, groupID int64) ([]model.GroupPlayer, error)
	}
}

// SimpleEventsHandler avoids complex dependency injection for now.
type SimpleEventsHandler struct {
	eventSvc  service.EventService
	draftSvc  service.DraftService
	leagueSvc service.LeagueService
}

func NewEventsHandler(
	eventSvc service.EventService,
	draftSvc service.DraftService,
	leagueSvc service.LeagueService,
) *SimpleEventsHandler {
	return &SimpleEventsHandler{
		eventSvc:  eventSvc,
		draftSvc:  draftSvc,
		leagueSvc: leagueSvc,
	}
}

// GET /api/v1/leagues/:id/events
func (h *SimpleEventsHandler) List(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}
	events, err := h.eventSvc.ListEvents(c.Request.Context(), leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// POST /api/v1/leagues/:id/events
func (h *SimpleEventsHandler) Create(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}
	var req struct {
		Title     string `json:"title"     binding:"required"`
		StartDate string `json:"startDate" binding:"required"`
		EndDate   string `json:"endDate"   binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid startDate format (use YYYY-MM-DD)"})
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endDate format (use YYYY-MM-DD)"})
		return
	}

	event, err := h.eventSvc.CreateDraftEvent(c.Request.Context(), leagueID, req.Title, start, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

// GET /api/v1/leagues/:id/events/:eid
func (h *SimpleEventsHandler) Get(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	detail, err := h.eventSvc.GetEventDetail(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

// POST /api/v1/leagues/:id/events/:eid/start
func (h *SimpleEventsHandler) Start(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	if err := h.eventSvc.StartEvent(c.Request.Context(), eventID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, _ := h.eventSvc.GetEvent(c.Request.Context(), eventID)
	c.JSON(http.StatusOK, event)
}

// POST /api/v1/leagues/:id/events/:eid/finish
func (h *SimpleEventsHandler) Finish(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	if err := h.draftSvc.FinishEvent(c.Request.Context(), eventID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, _ := h.eventSvc.GetEvent(c.Request.Context(), eventID)
	c.JSON(http.StatusOK, event)
}

// POST /api/v1/leagues/:id/events/:eid/next
func (h *SimpleEventsHandler) CreateNext(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	newEvent, err := h.draftSvc.CreateDraft(c.Request.Context(), leagueID, eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newEvent)
}

// PUT /api/v1/leagues/:id/events/:eid/config
func (h *SimpleEventsHandler) UpdateConfig(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	var cfg model.LeagueConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.draftSvc.RecreateDraft(c.Request.Context(), eventID, cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
