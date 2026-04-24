package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"league-api/internal/middleware"
	"league-api/internal/service"
)

type ProfileHandler struct {
	profileSvc service.ProfileService
}

func NewProfileHandler(profileSvc service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc}
}

// GET /secured/profile — own profile
func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	detail, err := h.profileSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

// PUT /secured/profile
func (h *ProfileHandler) UpsertMyProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req service.UpsertProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	detail, err := h.profileSvc.UpsertProfile(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

// GET /public/countries
func (h *ProfileHandler) ListCountries(c *gin.Context) {
	countries, err := h.profileSvc.ListCountries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, countries)
}

// GET /public/countries/:cid/cities
func (h *ProfileHandler) ListCities(c *gin.Context) {
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country id"})
		return
	}
	cities, err := h.profileSvc.ListCities(c.Request.Context(), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cities)
}

// POST /secured/countries/:cid/cities
func (h *ProfileHandler) AddCity(c *gin.Context) {
	cid, err := strconv.Atoi(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country id"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	city, err := h.profileSvc.AddCity(c.Request.Context(), req.Name, cid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, city)
}

// GET /public/blades
func (h *ProfileHandler) ListBlades(c *gin.Context) {
	blades, err := h.profileSvc.ListBlades(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, blades)
}

// POST /secured/blades
func (h *ProfileHandler) AddBlade(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	blade, err := h.profileSvc.AddBlade(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, blade)
}

// GET /public/rubbers
func (h *ProfileHandler) ListRubbers(c *gin.Context) {
	rubbers, err := h.profileSvc.ListRubbers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rubbers)
}

// POST /secured/rubbers
func (h *ProfileHandler) AddRubber(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rubber, err := h.profileSvc.AddRubber(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rubber)
}
