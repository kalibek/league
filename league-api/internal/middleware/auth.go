package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"league-api/internal/repository"
	"league-api/internal/service"
)

const (
	ContextUserID  = "userID"
	ContextRoles   = "roles"   // map[int64][]string: leagueID → role names
	ContextIsAdmin = "isAdmin" // bool
)

// Auth extracts and validates the JWT from Authorization header or session cookie.
// It loads the user's roles and attaches userID + roles to the Gin context.
// On failure it aborts with 401.
func Auth(authSvc service.AuthService, leagueRepo repository.LeagueRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		userID, err := authSvc.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Load roles for this user (all leagues).
		roles := loadUserRoles(c.Request.Context(), userID, leagueRepo)

		// Load isAdmin flag.
		isAdmin := false
		if u, err := authSvc.GetUser(c.Request.Context(), userID); err == nil {
			isAdmin = u.IsAdmin
		}

		c.Set(ContextUserID, userID)
		c.Set(ContextRoles, roles)
		c.Set(ContextIsAdmin, isAdmin)
		c.Next()
	}
}

// RequireAuth is a lightweight middleware that only checks for a valid JWT.
func RequireAuth(authSvc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		userID, err := authSvc.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(ContextUserID, userID)
		c.Next()
	}
}

// RequireAdmin aborts with 403 if the user is not an admin.
// Must be used after Auth middleware.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		c.Next()
	}
}

// IsAdmin returns true if the authenticated user has the admin flag.
func IsAdmin(c *gin.Context) bool {
	v, _ := c.Get(ContextIsAdmin)
	b, _ := v.(bool)
	return b
}

// RequireMaintainer aborts with 403 if the user is not a Maintainer for leagueID.
// Must be used after Auth middleware.
func RequireMaintainer(leagueID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasRole(c, leagueID, "maintainer") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "maintainer role required"})
			return
		}
		c.Next()
	}
}

// RequireUmpire aborts with 403 if the user is neither Umpire nor Maintainer.
func RequireUmpire(leagueID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasRole(c, leagueID, "umpire") && !hasRole(c, leagueID, "maintainer") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "umpire role required"})
			return
		}
		c.Next()
	}
}

// GetUserID retrieves the authenticated userID from context.
func GetUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// GetRoles retrieves the role map from context.
func GetRoles(c *gin.Context) map[int64][]string {
	v, _ := c.Get(ContextRoles)
	if v == nil {
		return map[int64][]string{}
	}
	roles, _ := v.(map[int64][]string)
	return roles
}

// ---

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := c.Cookie("session"); err == nil {
		return cookie
	}
	return ""
}

func hasRole(c *gin.Context, leagueID int64, role string) bool {
	roles := GetRoles(c)
	for _, r := range roles[leagueID] {
		if r == role {
			return true
		}
	}
	return false
}

// loadUserRoles fetches all UserRole records for a user in a single query
// and groups them by leagueID → role names. On any error it returns an empty map (non-fatal).
func loadUserRoles(ctx context.Context, userID int64, leagueRepo repository.LeagueRepository) map[int64][]string {
	urs, err := leagueRepo.GetAllUserRoles(ctx, userID)
	if err != nil {
		return map[int64][]string{}
	}
	result := make(map[int64][]string, len(urs))
	for _, ur := range urs {
		name := roleIDToName(ur.RoleID)
		result[ur.LeagueID] = append(result[ur.LeagueID], name)
	}
	return result
}

func roleIDToName(id int) string {
	switch id {
	case 1:
		return "player"
	case 2:
		return "umpire"
	case 3:
		return "maintainer"
	default:
		return "unknown"
	}
}
