package model

import "time"

type User struct {
	UserID        int64     `db:"user_id"        json:"userId"`
	FirstName     string    `db:"first_name"     json:"firstName"`
	LastName      string    `db:"last_name"      json:"lastName"`
	Email         string    `db:"email"          json:"email"`
	IsAdmin       bool      `db:"is_admin"       json:"isAdmin"`
	CurrentRating float64   `db:"current_rating" json:"currentRating"`
	Deviation     float64   `db:"deviation"      json:"deviation"`
	Volatility    float64   `db:"volatility"     json:"volatility"`
	Created       time.Time `db:"created"        json:"created"`
	LastUpdated   time.Time `db:"last_updated"   json:"lastUpdated"`
	PasswordHash  *string   `db:"password_hash"  json:"-"`
}

// OAuthAccount links a provider identity to a user; stored in user_oauth_accounts.
type OAuthAccount struct {
	AccountID   int64  `db:"account_id"`
	UserID      int64  `db:"user_id"`
	Provider    string `db:"provider"`     // "google" | "facebook" | "apple"
	ProviderSub string `db:"provider_sub"` // subject ID from the provider
}

type Role struct {
	RoleID   int    `db:"role_id"   json:"roleId"`
	RoleName string `db:"role_name" json:"roleName"`
}

type UserRole struct {
	UserID   int64 `db:"user_id"   json:"userId"`
	RoleID   int   `db:"role_id"   json:"roleId"`
	LeagueID int64 `db:"league_id" json:"leagueId"`
}
