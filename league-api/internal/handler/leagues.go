package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"league-api/internal/middleware"
	"league-api/internal/model"
	"league-api/internal/service"
)

type LeaguesHandler struct {
	leagueSvc service.LeagueService
}

func NewLeaguesHandler(leagueSvc service.LeagueService) *LeaguesHandler {
	return &LeaguesHandler{leagueSvc}
}

// GET /api/v1/leagues
func (h *LeaguesHandler) List(c *gin.Context) {
	summaries, err := h.leagueSvc.ListLeagueSummaries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summaries)
}

// POST /api/v1/leagues
func (h *LeaguesHandler) Create(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		return
	}

	var req struct {
		Title         string             `json:"title"         binding:"required"`
		Description   string             `json:"description"`
		Configuration model.LeagueConfig `json:"configuration"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	league, err := h.leagueSvc.CreateLeague(c.Request.Context(), userID, req.Title, req.Description, req.Configuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, league)
}

// GET /api/v1/leagues/:id
func (h *LeaguesHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	league, err := h.leagueSvc.GetLeague(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, league)
}

// PUT /api/v1/leagues/:id/config
func (h *LeaguesHandler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var cfg model.LeagueConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.leagueSvc.UpdateConfig(c.Request.Context(), id, cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/v1/leagues/:id/roles
func (h *LeaguesHandler) AssignRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	callerID, _ := middleware.GetUserID(c)
	if !middleware.IsAdmin(c) && !h.leagueSvc.IsMaintainer(c.Request.Context(), id, callerID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or maintainer role required"})
		return
	}

	var req struct {
		UserID   int64  `json:"userId"   binding:"required"`
		RoleName string `json:"roleName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.leagueSvc.AssignRole(c.Request.Context(), id, req.UserID, req.RoleName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /api/v1/leagues/:id/roles
func (h *LeaguesHandler) RemoveRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	callerID, _ := middleware.GetUserID(c)
	if !middleware.IsAdmin(c) && !h.leagueSvc.IsMaintainer(c.Request.Context(), id, callerID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or maintainer role required"})
		return
	}

	var req struct {
		UserID   int64  `json:"userId"   binding:"required"`
		RoleName string `json:"roleName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.leagueSvc.RemoveRole(c.Request.Context(), id, req.UserID, req.RoleName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/v1/leagues/:id/roles
func (h *LeaguesHandler) GetRoles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	callerID, _ := middleware.GetUserID(c)
	if !middleware.IsAdmin(c) && !h.leagueSvc.IsMaintainer(c.Request.Context(), id, callerID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin or maintainer role required"})
		return
	}

	roles, err := h.leagueSvc.ListLeagueRoles(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}
