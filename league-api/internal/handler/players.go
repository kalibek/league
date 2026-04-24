package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"league-api/internal/service"
)

type PlayersHandler struct {
	playerSvc service.PlayerService
}

func NewPlayersHandler(playerSvc service.PlayerService) *PlayersHandler {
	return &PlayersHandler{playerSvc: playerSvc}
}

// GET /api/v1/players?q=&sort=&limit=&offset=
func (h *PlayersHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sort := c.DefaultQuery("sort", "rating")
	q := c.Query("q")

	players, err := h.playerSvc.ListPlayers(c.Request.Context(), q, limit, offset, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, players)
}

// POST /api/v1/players
func (h *PlayersHandler) Create(c *gin.Context) {
	var req struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName"  binding:"required"`
		Email     string `json:"email"     binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player, err := h.playerSvc.CreatePlayer(c.Request.Context(), req.FirstName, req.LastName, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, player)
}

// GET /api/v1/players/:id
func (h *PlayersHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	profile, err := h.playerSvc.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// GET /api/v1/players/:id/events?limit=5&offset=0
func (h *PlayersHandler) ListEvents(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	page, err := h.playerSvc.GetPlayerEvents(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, page)
}

// POST /api/v1/players/import
func (h *PlayersHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
		return
	}
	defer file.Close()

	result, err := h.playerSvc.ImportCSV(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Flatten errors to string slice for API response.
	errStrings := make([]string, 0, len(result.Errors))
	for _, e := range result.Errors {
		errStrings = append(errStrings, "row "+strconv.Itoa(e.Row)+": "+e.Message)
	}

	c.JSON(http.StatusOK, gin.H{
		"imported": result.Imported,
		"skipped":  result.Skipped,
		"errors":   errStrings,
	})
}
