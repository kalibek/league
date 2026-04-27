package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"league-api/internal/repository"
	"league-api/internal/service"
)

type GroupsHandler struct {
	groupSvc   service.GroupService
	draftSvc   service.DraftService
	matchSvc   service.MatchService
	leagueRepo repository.LeagueRepository
	eventRepo  repository.EventRepository
}

func NewGroupsHandler(
	groupSvc service.GroupService,
	draftSvc service.DraftService,
	matchSvc service.MatchService,
	leagueRepo repository.LeagueRepository,
	eventRepo repository.EventRepository,
) *GroupsHandler {
	return &GroupsHandler{
		groupSvc:   groupSvc,
		draftSvc:   draftSvc,
		matchSvc:   matchSvc,
		leagueRepo: leagueRepo,
		eventRepo:  eventRepo,
	}
}

// GET /api/v1/public/events/:eid/groups
func (h *GroupsHandler) List(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	groups, err := h.groupSvc.ListGroups(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// POST /api/v1/secured/events/:eid/groups
func (h *GroupsHandler) Create(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	var req struct {
		Division  string `json:"division"  binding:"required"`
		GroupNo   int    `json:"groupNo"   binding:"required"`
		Scheduled string `json:"scheduled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	scheduled, err := time.Parse(time.RFC3339, req.Scheduled)
	if err != nil {
		// Try date-only.
		scheduled, err = time.Parse("2006-01-02", req.Scheduled)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduled format (use RFC3339 or YYYY-MM-DD)"})
			return
		}
	}
	grp, err := h.groupSvc.CreateGroup(c.Request.Context(), eventID, req.Division, req.GroupNo, scheduled)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, grp)
}

// POST /api/v1/secured/events/:eid/groups/:gid/seed
func (h *GroupsHandler) SeedPlayer(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		UserID int64 `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupSvc.SeedPlayer(c.Request.Context(), groupID, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// DELETE /api/v1/secured/events/:eid/groups/:gid/players/:gpid
func (h *GroupsHandler) RemovePlayer(c *gin.Context) {
	groupPlayerID, err := strconv.ParseInt(c.Param("gpid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group player id"})
		return
	}
	if err := h.groupSvc.RemovePlayer(c.Request.Context(), groupPlayerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/v1/events/:eid/groups/:gid
func (h *GroupsHandler) Get(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	grp, players, matches, err := h.groupSvc.GetGroupDetail(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groupId":   grp.GroupID,
		"eventId":   grp.EventID,
		"status":    grp.Status,
		"division":  grp.Division,
		"groupNo":   grp.GroupNo,
		"scheduled": grp.Scheduled,
		"players":   players,
		"matches":   matches,
	})
}

// POST /api/v1/events/:eid/groups/:gid/finish
func (h *GroupsHandler) Finish(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}

	if err := h.draftSvc.FinishGroup(c.Request.Context(), groupID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/v1/secured/events/:eid/groups/:gid/reopen
func (h *GroupsHandler) Reopen(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	if err := h.draftSvc.ReopenGroup(c.Request.Context(), groupID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/v1/events/:eid/groups/:gid/players
func (h *GroupsHandler) AddPlayer(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		UserID int64 `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupSvc.AddNonCalculatedPlayer(c.Request.Context(), groupID, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// PUT /api/v1/events/:eid/groups/:gid/players/:pid/place
func (h *GroupsHandler) SetManualPlace(c *gin.Context) {
	groupPlayerID, err := strconv.ParseInt(c.Param("pid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}
	var req struct {
		Place int16 `json:"place" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.groupSvc.SetManualPlace(c.Request.Context(), groupPlayerID, req.Place); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PUT /api/v1/events/:eid/groups/:gid/placement
func (h *GroupsHandler) SetManualPlacements(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	var req struct {
		OrderedPlayerIDs []int64 `json:"orderedPlayerIds" binding:"required,min=2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.draftSvc.SetManualPlacements(c.Request.Context(), groupID, req.OrderedPlayerIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// getGamesToWin fetches gamesToWin from league config via group → event → league chain.
func (h *GroupsHandler) getGamesToWin(ctx context.Context, groupID int64) int {
	// This requires getting group → eventID → leagueID → config.
	// We call the group detail to get eventID.
	grp, _, _, err := h.groupSvc.GetGroupDetail(ctx, groupID)
	if err != nil {
		return 3 // safe default
	}
	ev, err := h.eventRepo.GetByID(ctx, grp.EventID)
	if err != nil {
		return 3
	}
	league, err := h.leagueRepo.GetByID(ctx, ev.LeagueID)
	if err != nil {
		return 3
	}
	if league.Config.GamesToWin <= 0 {
		return 3
	}
	return league.Config.GamesToWin
}
