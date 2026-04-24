package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"league-api/internal/middleware"
	"league-api/internal/repository"
	"league-api/internal/service"
)

type AuthHandler struct {
	authSvc     service.AuthService
	leagueRepo  repository.LeagueRepository
	frontendURL string
}

func NewAuthHandler(authSvc service.AuthService, leagueRepo repository.LeagueRepository, frontendURL string) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, leagueRepo: leagueRepo, frontendURL: frontendURL}
}

// GET /auth/login?provider=google|facebook|apple
func (h *AuthHandler) Login(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		provider = "google"
	}

	state, err := service.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "state generation failed"})
		return
	}

	// Store state in cookie for validation on callback.
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	url, err := h.authSvc.GetAuthURL(provider, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, url)
}

// GET /auth/callback?code=...&state=...&provider=...
func (h *AuthHandler) Callback(c *gin.Context) {
	provider := c.Query("provider")
	code := c.Query("code")
	state := c.Query("state")

	// Validate state.
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	_, jwtToken, err := h.authSvc.HandleCallback(c.Request.Context(), provider, code, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set JWT in session cookie.
	c.SetCookie("session", jwtToken, 30*24*3600, "/", "", false, true)
	c.Redirect(http.StatusFound, h.frontendURL)
}

// POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var body struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName"  binding:"required"`
		Email     string `json:"email"     binding:"required,email"`
		Password  string `json:"password"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, jwtToken, err := h.authSvc.Register(c.Request.Context(), body.FirstName, body.LastName, body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("session", jwtToken, 30*24*3600, "/", "", false, true)
	c.JSON(http.StatusCreated, user)
}

// POST /auth/login/email
func (h *AuthHandler) EmailLogin(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, jwtToken, err := h.authSvc.EmailLogin(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("session", jwtToken, 30*24*3600, "/", "", false, true)
	c.JSON(http.StatusOK, user)
}

// GET /auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	roles := middleware.GetRoles(c)

	// Convert roles map to a flat list for the frontend.
	type roleEntry struct {
		UserID    int64  `json:"userId"`
		LeagueID  int64  `json:"leagueId"`
		RoleName  string `json:"roleName"`
	}
	var roleList []roleEntry
	for leagueID, names := range roles {
		for _, name := range names {
			roleList = append(roleList, roleEntry{
				UserID:   userID,
				LeagueID: leagueID,
				RoleName: name,
			})
		}
	}

	user, err := h.authSvc.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":        user.UserID,
		"firstName":     user.FirstName,
		"lastName":      user.LastName,
		"email":         user.Email,
		"isAdmin":       user.IsAdmin,
		"currentRating": user.CurrentRating,
		"deviation":     user.Deviation,
		"volatility":    user.Volatility,
		"roles":         roleList,
	})
}
