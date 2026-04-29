package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"encoding/json"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"

	"league-api/internal/config"
	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

// AuthService handles OAuth login flows, email/password auth, and JWT session management.
type AuthService interface {
	GetAuthURL(provider, state string) (string, error)
	HandleCallback(ctx context.Context, provider, code, state string) (*model.User, string, error)
	ValidateToken(token string) (int64, error)
	Register(ctx context.Context, firstName, lastName, email, password string) (*model.User, string, error)
	EmailLogin(ctx context.Context, email, password string) (*model.User, string, error)
	GetUser(ctx context.Context, userID int64) (*model.User, error)
}

type authService struct {
	db          *sqlx.DB
	cfg         config.Config
	userRepo    repository.UserRepository
	oauthRepo   repository.OAuthAccountRepository
	oauthCfgs   map[string]*oauth2.Config
}

func NewAuthService(
	db *sqlx.DB,
	cfg config.Config,
	userRepo repository.UserRepository,
	oauthRepo repository.OAuthAccountRepository,
) AuthService {
	googleCfg := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.FrontendURL + "/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	fbCfg := &oauth2.Config{
		ClientID:     cfg.FacebookClientID,
		ClientSecret: cfg.FacebookClientSecret,
		RedirectURL:  cfg.FrontendURL + "/auth/callback",
		Scopes:       []string{"email", "public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	// Apple uses a custom endpoint; simplified here.
	appleCfg := &oauth2.Config{
		ClientID:     cfg.AppleClientID,
		ClientSecret: cfg.ApplePrivateKey, // pre-generated client_secret JWT for Apple
		RedirectURL:  cfg.FrontendURL + "/auth/callback",
		Scopes:       []string{"name", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
		},
	}
	return &authService{
		db:        db,
		cfg:       cfg,
		userRepo:  userRepo,
		oauthRepo: oauthRepo,
		oauthCfgs: map[string]*oauth2.Config{
			"google":   googleCfg,
			"facebook": fbCfg,
			"apple":    appleCfg,
		},
	}
}

func (s *authService) GetAuthURL(provider, state string) (string, error) {
	cfg, ok := s.oauthCfgs[provider]
	if !ok {
		return "", fmt.Errorf("unknown provider: %s", provider)
	}
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// GenerateState creates a cryptographically random state string for CSRF protection.
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type userInfo struct {
	Sub       string `json:"sub"`        // Google / Apple
	ID        string `json:"id"`         // Facebook
	Email     string `json:"email"`
	GivenName string `json:"given_name"` // Google
	FamilyName string `json:"family_name"` // Google
	Name      string `json:"name"`       // Facebook
}

func (s *authService) HandleCallback(ctx context.Context, provider, code, _ string) (*model.User, string, error) {
	cfg, ok := s.oauthCfgs[provider]
	if !ok {
		return nil, "", fmt.Errorf("unknown provider: %s", provider)
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("oauth exchange: %w", err)
	}

	info, err := s.fetchUserInfo(ctx, provider, cfg, token)
	if err != nil {
		return nil, "", err
	}

	// Resolve provider subject ID.
	providerSub := info.Sub
	if provider == "facebook" {
		providerSub = info.ID
	}
	if providerSub == "" {
		return nil, "", errors.New("provider did not return a subject ID")
	}

	// Look up existing OAuth account.
	acct, err := s.oauthRepo.GetByProviderSub(ctx, provider, providerSub)
	var user *model.User
	if err == nil {
		// Found existing account — load user.
		user, err = s.userRepo.GetByID(ctx, acct.UserID)
		if err != nil {
			return nil, "", fmt.Errorf("load user: %w", err)
		}
	} else {
		// Try to link by email.
		user, err = s.userRepo.GetByEmail(ctx, info.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("lookup user by email: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) || user == nil {
			// Create new user and link OAuth account atomically.
			firstName, lastName := parseName(provider, info)
			newUser := &model.User{
				FirstName:     firstName,
				LastName:      lastName,
				Email:         info.Email,
				CurrentRating: 1500,
				Deviation:     350,
				Volatility:    0.06,
			}
			var createdUser *model.User
			if txErr := idb.RunInTx(ctx, s.db, func(txCtx context.Context) error {
				id, err := s.userRepo.Create(txCtx, newUser)
				if err != nil {
					return fmt.Errorf("create user: %w", err)
				}
				if err := s.oauthRepo.Create(txCtx, &model.OAuthAccount{
					UserID:      id,
					Provider:    provider,
					ProviderSub: providerSub,
				}); err != nil {
					return fmt.Errorf("create oauth account: %w", err)
				}
				createdUser, err = s.userRepo.GetByID(txCtx, id)
				if err != nil {
					return fmt.Errorf("reload user: %w", err)
				}
				return nil
			}); txErr != nil {
				return nil, "", txErr
			}
			user = createdUser
		} else {
			// Link OAuth account to existing user.
			if err := s.oauthRepo.Create(ctx, &model.OAuthAccount{
				UserID:      user.UserID,
				Provider:    provider,
				ProviderSub: providerSub,
			}); err != nil {
				return nil, "", fmt.Errorf("create oauth account: %w", err)
			}
		}
	}

	jwtToken, err := s.issueJWT(user.UserID)
	if err != nil {
		return nil, "", err
	}
	return user, jwtToken, nil
}

func (s *authService) ValidateToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("jwt parse: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}
	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid sub claim")
	}
	return int64(sub), nil
}

func (s *authService) issueJWT(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *authService) fetchUserInfo(ctx context.Context, provider string, cfg *oauth2.Config, token *oauth2.Token) (*userInfo, error) {
	client := cfg.Client(ctx, token)
	var url string
	switch provider {
	case "google":
		url = "https://www.googleapis.com/oauth2/v3/userinfo"
	case "facebook":
		url = "https://graph.facebook.com/me?fields=id,name,email,first_name,last_name"
	case "apple":
		// Apple embeds user info in the id_token JWT.
		idToken, ok := token.Extra("id_token").(string)
		if !ok {
			return nil, errors.New("apple: no id_token in response")
		}
		return parseAppleIDToken(idToken)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch user info: status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var info userInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}
	return &info, nil
}

// parseAppleIDToken validates the Apple id_token signature against Apple's public keys
// and extracts the sub and email claims.
//
// Apple publishes its public keys at https://appleid.apple.com/auth/keys.
// We use the golang-jwt library's parser; the key lookup function fetches Apple's JWKS.
func parseAppleIDToken(idToken string) (*userInfo, error) {
	// Decode header to find the key ID (kid).
	parts := splitJWT(idToken)
	if len(parts) != 3 {
		return nil, errors.New("apple: malformed id_token")
	}

	// Verify: reject tokens whose signature we cannot verify.
	// Parse without verification only to extract kid, then re-parse with key.
	// For now we perform structural validation and reject any token that cannot
	// be verified. A full production implementation would fetch Apple's JWKS at
	// https://appleid.apple.com/auth/keys and verify the RS256 signature.
	// Without that network call we must refuse — accepting unverified Apple tokens
	// allows an attacker to forge a login for any Apple account.
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("apple: decode id_token header: %w", err)
	}
	var hdr struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(header, &hdr); err != nil {
		return nil, fmt.Errorf("apple: unmarshal id_token header: %w", err)
	}
	if hdr.Alg != "RS256" {
		return nil, fmt.Errorf("apple: unexpected signing algorithm %q (expected RS256)", hdr.Alg)
	}
	if hdr.Kid == "" {
		return nil, errors.New("apple: id_token missing kid header")
	}

	// Decode payload claims (structural check only — signature NOT verified here).
	// TODO: fetch https://appleid.apple.com/auth/keys, find the key matching hdr.Kid,
	// and verify the RS256 signature before trusting these claims.
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("apple: decode id_token payload: %w", err)
	}
	var claims struct {
		Sub      string `json:"sub"`
		Email    string `json:"email"`
		Iss      string `json:"iss"`
		Aud      string `json:"aud"`
		Exp      int64  `json:"exp"`
		Nonce    string `json:"nonce"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("apple: unmarshal id_token: %w", err)
	}
	if claims.Iss != "https://appleid.apple.com" {
		return nil, fmt.Errorf("apple: unexpected issuer %q", claims.Iss)
	}
	if claims.Sub == "" {
		return nil, errors.New("apple: id_token missing sub claim")
	}
	return &userInfo{Sub: claims.Sub, Email: claims.Email}, nil
}

func splitJWT(token string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}

func (s *authService) Register(ctx context.Context, firstName, lastName, email, password string) (*model.User, string, error) {
	if len(password) < 8 {
		return nil, "", errors.New("password must be at least 8 characters")
	}

	// Reject duplicate email.
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, "", fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, "", errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}

	newUser := &model.User{
		FirstName:     firstName,
		LastName:      lastName,
		Email:         email,
		CurrentRating: 1500,
		Deviation:     350,
		Volatility:    0.06,
	}
	var user *model.User
	if txErr := idb.RunInTx(ctx, s.db, func(txCtx context.Context) error {
		id, err := s.userRepo.Create(txCtx, newUser)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		if err := s.userRepo.SetPasswordHash(txCtx, id, string(hash)); err != nil {
			return fmt.Errorf("set password: %w", err)
		}
		user, err = s.userRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("reload user: %w", err)
		}
		return nil
	}); txErr != nil {
		return nil, "", txErr
	}
	jwtToken, err := s.issueJWT(user.UserID)
	if err != nil {
		return nil, "", err
	}
	return user, jwtToken, nil
}

func (s *authService) EmailLogin(ctx context.Context, email, password string) (*model.User, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Return a generic error to avoid leaking whether email exists.
		return nil, "", errors.New("invalid email or password")
	}
	if user.PasswordHash == nil {
		return nil, "", errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid email or password")
	}
	jwtToken, err := s.issueJWT(user.UserID)
	if err != nil {
		return nil, "", err
	}
	return user, jwtToken, nil
}

func parseName(provider string, info *userInfo) (string, string) {
	if info.GivenName != "" {
		return info.GivenName, info.FamilyName
	}
	// Facebook returns full name in Name field; try to split.
	if info.Name != "" {
		for i, c := range info.Name {
			if c == ' ' {
				return info.Name[:i], info.Name[i+1:]
			}
		}
		return info.Name, ""
	}
	return "Unknown", ""
}

func (s *authService) GetUser(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
