package service

import (
	"context"
	"database/sql"
	"testing"

	"league-api/internal/config"
	"league-api/internal/model"
)

// authMockUserRepo is a minimal user repo for auth service tests.
type authMockUserRepo struct {
	users   map[string]*model.User // keyed by email
	byID    map[int64]*model.User
	getErr  error
}

func (m *authMockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if u, ok := m.byID[id]; ok {
		return u, nil
	}
	return nil, errFromString("not found")
}

func (m *authMockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	// Return sql.ErrNoRows so Register interprets it as "not found" (no duplicate).
	return nil, sql.ErrNoRows
}

func (m *authMockUserRepo) Create(ctx context.Context, u *model.User) (int64, error) { return 1, nil }

func (m *authMockUserRepo) List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *authMockUserRepo) Search(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *authMockUserRepo) UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error {
	return nil
}

func (m *authMockUserRepo) ResetAllRatings(ctx context.Context) error { return nil }

func (m *authMockUserRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	return nil
}

func (m *authMockUserRepo) UpdateName(ctx context.Context, userID int64, firstName, lastName string) error {
	return nil
}

// authMockOAuthRepo is a minimal oauth repo for auth service tests.
type authMockOAuthRepo struct{}

func (m *authMockOAuthRepo) GetByProviderSub(ctx context.Context, provider, sub string) (*model.OAuthAccount, error) {
	return nil, errFromString("not found")
}

func (m *authMockOAuthRepo) Create(ctx context.Context, a *model.OAuthAccount) error { return nil }

func (m *authMockOAuthRepo) ListByUser(ctx context.Context, userID int64) ([]model.OAuthAccount, error) {
	return nil, nil
}

// buildTestAuthService creates an authService with a test JWT secret.
func buildTestAuthService(ur *authMockUserRepo) *authService {
	cfg := config.Config{
		JWTSecret:   "test-secret-key",
		FrontendURL: "http://localhost:5173",
	}
	svc := NewAuthService(cfg, ur, &authMockOAuthRepo{})
	return svc.(*authService)
}

// --- GetAuthURL tests ---

func TestGetAuthURL_UnknownProvider(t *testing.T) {
	ur := &authMockUserRepo{users: map[string]*model.User{}, byID: map[int64]*model.User{}}
	svc := buildTestAuthService(ur)

	_, err := svc.GetAuthURL("unknown_provider", "state123")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestGetAuthURL_Google(t *testing.T) {
	ur := &authMockUserRepo{users: map[string]*model.User{}, byID: map[int64]*model.User{}}
	svc := buildTestAuthService(ur)

	url, err := svc.GetAuthURL("google", "state123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty auth URL")
	}
}

func TestGetAuthURL_Facebook(t *testing.T) {
	ur := &authMockUserRepo{users: map[string]*model.User{}, byID: map[int64]*model.User{}}
	svc := buildTestAuthService(ur)

	url, err := svc.GetAuthURL("facebook", "state456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty auth URL")
	}
}

func TestGetAuthURL_Apple(t *testing.T) {
	ur := &authMockUserRepo{users: map[string]*model.User{}, byID: map[int64]*model.User{}}
	svc := buildTestAuthService(ur)

	url, err := svc.GetAuthURL("apple", "state789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty auth URL")
	}
}

// --- GenerateState test ---

func TestGenerateState(t *testing.T) {
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(state) == 0 {
		t.Error("expected non-empty state string")
	}
}

// --- ValidateToken tests ---

func TestValidateToken_Valid(t *testing.T) {
	ur := &authMockUserRepo{users: map[string]*model.User{}, byID: map[int64]*model.User{}}
	svc := buildTestAuthService(ur)

	// Issue a real JWT then validate it.
	tokenStr, err := svc.issueJWT(42)
	if err != nil {
		t.Fatalf("unexpected error issuing JWT: %v", err)
	}

	userID, err := svc.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error validating token: %v", err)
	}
	if userID != 42 {
		t.Errorf("expected userID 42, got %d", userID)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	ur := &authMockUserRepo{users: map[string]*model.User{}, byID: map[int64]*model.User{}}
	svc := buildTestAuthService(ur)

	_, err := svc.ValidateToken("not-a-valid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	ur := &authMockUserRepo{users: map[string]*model.User{}, byID: map[int64]*model.User{}}
	svc1 := buildTestAuthService(ur)

	// Issue with one secret.
	tokenStr, err := svc1.issueJWT(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate with different secret.
	cfg2 := config.Config{JWTSecret: "different-secret", FrontendURL: "http://localhost:5173"}
	svc2 := NewAuthService(cfg2, ur, &authMockOAuthRepo{}).(*authService)

	_, err = svc2.ValidateToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for token with wrong secret")
	}
}

// --- GetUser tests ---

func TestGetUser_Found(t *testing.T) {
	ur := &authMockUserRepo{
		users: map[string]*model.User{},
		byID:  map[int64]*model.User{7: {UserID: 7, FirstName: "Alice"}},
	}
	svc := buildTestAuthService(ur)

	user, err := svc.GetUser(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != 7 {
		t.Errorf("expected userID 7, got %d", user.UserID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	ur := &authMockUserRepo{
		users:  map[string]*model.User{},
		byID:   map[int64]*model.User{},
		getErr: errFromString("not found"),
	}
	svc := buildTestAuthService(ur)

	_, err := svc.GetUser(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

// --- parseName tests ---

func TestParseName_WithGivenName(t *testing.T) {
	info := &userInfo{GivenName: "Alice", FamilyName: "Smith"}
	first, last := parseName("google", info)
	if first != "Alice" || last != "Smith" {
		t.Errorf("expected Alice Smith, got %s %s", first, last)
	}
}

func TestParseName_WithFullName(t *testing.T) {
	info := &userInfo{Name: "Bob Jones"}
	first, last := parseName("facebook", info)
	if first != "Bob" || last != "Jones" {
		t.Errorf("expected Bob Jones, got %s %s", first, last)
	}
}

func TestParseName_WithNameNoSpace(t *testing.T) {
	info := &userInfo{Name: "Cher"}
	first, last := parseName("facebook", info)
	if first != "Cher" || last != "" {
		t.Errorf("expected Cher '', got %s %s", first, last)
	}
}

func TestParseName_Empty(t *testing.T) {
	info := &userInfo{}
	first, last := parseName("google", info)
	if first != "Unknown" || last != "" {
		t.Errorf("expected Unknown '', got %s %s", first, last)
	}
}

// --- EmailLogin tests ---

func TestEmailLogin_UserNotFound(t *testing.T) {
	ur := &authMockUserRepo{
		users: map[string]*model.User{},
		byID:  map[int64]*model.User{},
	}
	svc := buildTestAuthService(ur)

	_, _, err := svc.EmailLogin(context.Background(), "nobody@example.com", "password")
	if err == nil {
		t.Fatal("expected error for unknown email")
	}
}

func TestEmailLogin_NilPasswordHash(t *testing.T) {
	ur := &authMockUserRepo{
		users: map[string]*model.User{
			"alice@example.com": {UserID: 1, Email: "alice@example.com", PasswordHash: nil},
		},
		byID: map[int64]*model.User{},
	}
	svc := buildTestAuthService(ur)

	_, _, err := svc.EmailLogin(context.Background(), "alice@example.com", "password")
	if err == nil {
		t.Fatal("expected error when PasswordHash is nil")
	}
}

func TestEmailLogin_WrongPassword(t *testing.T) {
	// bcrypt hash of "correct_password".
	hashStr := "$2a$10$YgE4i3HkVPGBRPMJjJ4s3.LcY4bH9g7vwK4OJJ4VjUkWg7IrKfuBy"
	ur := &authMockUserRepo{
		users: map[string]*model.User{
			"alice@example.com": {UserID: 1, Email: "alice@example.com", PasswordHash: &hashStr},
		},
		byID: map[int64]*model.User{},
	}
	svc := buildTestAuthService(ur)

	_, _, err := svc.EmailLogin(context.Background(), "alice@example.com", "wrong_password")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}
