package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"league-api/internal/service"
)

type mergeRequest struct {
	TargetID  int64   `json:"targetId"  binding:"required"`
	SourceIDs []int64 `json:"sourceIds" binding:"required,min=1"`
}

type AdminHandler struct {
	ratingSvc service.RatingService
	playerSvc service.PlayerService
}

func NewAdminHandler(ratingSvc service.RatingService, playerSvc service.PlayerService) *AdminHandler {
	return &AdminHandler{ratingSvc: ratingSvc, playerSvc: playerSvc}
}

// GET /api/v1/admin/players/duplicates
func (h *AdminHandler) GetDuplicates(c *gin.Context) {
	groups, err := h.playerSvc.FindDuplicates(c.Request.Context())
	if err != nil {
		log.Printf("[handler] AdminHandler.GetDuplicates: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// POST /api/v1/admin/players/merge
func (h *AdminHandler) MergePlayers(c *gin.Context) {
	var req mergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.playerSvc.MergeUsers(c.Request.Context(), req.TargetID, req.SourceIDs)
	if err != nil {
		log.Printf("[handler] AdminHandler.MergePlayers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// POST /api/v1/admin/ratings/recalculate
// Wipes all rating history, resets all user ratings to initial Glicko2 params,
// then replays every DONE event in chronological order using per-match Glicko2.
func (h *AdminHandler) RecalculateRatings(c *gin.Context) {
	result, err := h.ratingSvc.RecalculateAllRatings(c.Request.Context())
	if err != nil {
		log.Printf("[handler] AdminHandler.RecalculateRatings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":               true,
		"eventsProcessed":  result.EventsProcessed,
		"groupsProcessed":  result.GroupsProcessed,
		"matchesProcessed": result.MatchesProcessed,
	})
}
