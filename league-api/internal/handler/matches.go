package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"league-api/internal/repository"
	"league-api/internal/service"
)

type MatchesHandler struct {
	matchSvc   service.MatchService
	groupSvc   service.GroupService
	leagueRepo repository.LeagueRepository
	eventRepo  repository.EventRepository
}

func NewMatchesHandler(
	matchSvc service.MatchService,
	groupSvc service.GroupService,
	leagueRepo repository.LeagueRepository,
	eventRepo repository.EventRepository,
) *MatchesHandler {
	return &MatchesHandler{
		matchSvc:   matchSvc,
		groupSvc:   groupSvc,
		leagueRepo: leagueRepo,
		eventRepo:  eventRepo,
	}
}

// PUT /api/v1/groups/:gid/matches/:mid
func (h *MatchesHandler) UpdateScore(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	matchID, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	var req struct {
		Score1    int16 `json:"score1"`
		Score2    int16 `json:"score2"`
		Withdraw1 bool  `json:"withdraw1"`
		Withdraw2 bool  `json:"withdraw2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gamesToWin := h.getGamesToWin(c.Request.Context(), groupID)

	if err := h.matchSvc.UpdateScore(c.Request.Context(), matchID, req.Score1, req.Score2, gamesToWin, req.Withdraw1, req.Withdraw2); err != nil {
		log.Printf("[handler] MatchesHandler.UpdateScore: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MatchesHandler) getGamesToWin(ctx context.Context, groupID int64) int {
	grp, _, _, err := h.groupSvc.GetGroupDetail(ctx, groupID)
	if err != nil {
		return 3
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

// DELETE /api/v1/secured/groups/:gid/matches/:mid/score
func (h *MatchesHandler) ResetScore(c *gin.Context) {
	matchID, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}
	if err := h.matchSvc.ResetScore(c.Request.Context(), matchID); err != nil {
		log.Printf("[handler] MatchesHandler.ResetScore: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PUT /api/v1/groups/:gid/matches/:mid/table
func (h *MatchesHandler) SetTableNumber(c *gin.Context) {
	groupID, err := strconv.ParseInt(c.Param("gid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group id"})
		return
	}
	matchID, err := strconv.ParseInt(c.Param("mid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	var req struct {
		TableNumber int `json:"tableNumber"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventID := h.getEventIDForGroup(c.Request.Context(), groupID)
	if err := h.matchSvc.SetTableNumber(c.Request.Context(), matchID, req.TableNumber, eventID); err != nil {
		log.Printf("[handler] MatchesHandler.SetTableNumber: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/v1/events/:eid/tables-in-use
func (h *MatchesHandler) GetTablesInUse(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("eid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	tables, err := h.matchSvc.ListInProgressByEvent(c.Request.Context(), eventID)
	if err != nil {
		log.Printf("[handler] MatchesHandler.GetTablesInUse: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tablesInUse": tables})
}

func (h *MatchesHandler) getEventIDForGroup(ctx context.Context, groupID int64) int64 {
	grp, _, _, err := h.groupSvc.GetGroupDetail(ctx, groupID)
	if err != nil {
		return 0
	}
	return grp.EventID
}
