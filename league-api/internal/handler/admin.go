package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"league-api/internal/service"
)

type AdminHandler struct {
	ratingSvc service.RatingService
}

func NewAdminHandler(ratingSvc service.RatingService) *AdminHandler {
	return &AdminHandler{ratingSvc: ratingSvc}
}

// POST /api/v1/admin/ratings/recalculate
// Wipes all rating history, resets all user ratings to initial Glicko2 params,
// then replays every DONE event in chronological order using per-match Glicko2.
func (h *AdminHandler) RecalculateRatings(c *gin.Context) {
	result, err := h.ratingSvc.RecalculateAllRatings(c.Request.Context())
	if err != nil {
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
