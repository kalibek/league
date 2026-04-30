package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"league-api/internal/model"
)

// --- player service mocks ---

type psUserRepo struct {
	users       map[int64]*model.User
	byEmail     map[string]*model.User
	createErr   error
	nextID      int64
}

func newPSUserRepo() *psUserRepo {
	return &psUserRepo{
		users:   make(map[int64]*model.User),
		byEmail: make(map[string]*model.User),
		nextID:  1,
	}
}

func (m *psUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}

func (m *psUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if u, ok := m.byEmail[email]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}

func (m *psUserRepo) Create(ctx context.Context, u *model.User) (int64, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	id := m.nextID
	m.nextID++
	u.UserID = id
	m.users[id] = u
	m.byEmail[u.Email] = u
	return id, nil
}

func (m *psUserRepo) List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error) {
	result := make([]model.User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, *u)
	}
	return result, nil
}

func (m *psUserRepo) Search(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error) {
	var result []model.User
	for _, u := range m.users {
		if strings.Contains(u.FirstName, q) || strings.Contains(u.LastName, q) {
			result = append(result, *u)
		}
	}
	return result, nil
}

func (m *psUserRepo) UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error {
	return nil
}

func (m *psUserRepo) ResetAllRatings(ctx context.Context) error { return nil }

func (m *psUserRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error { return nil }

func (m *psUserRepo) UpdateName(ctx context.Context, userID int64, firstName, lastName string) error {
	return nil
}

type psRatingRepo struct {
	history map[int64][]model.RatingHistory // userID → history
	eventDelta map[int64]map[int64]float64  // userID → eventID → delta
}

func (m *psRatingRepo) InsertHistory(ctx context.Context, rh *model.RatingHistory) error {
	m.history[rh.UserID] = append(m.history[rh.UserID], *rh)
	return nil
}

func (m *psRatingRepo) GetByUser(ctx context.Context, userID int64) ([]model.RatingHistory, error) {
	return m.history[userID], nil
}

func (m *psRatingRepo) GetByUserInEvent(ctx context.Context, userID, eventID int64) ([]model.RatingHistory, error) {
	return nil, nil
}

func (m *psRatingRepo) DeleteByGroup(ctx context.Context, groupID int64) error { return nil }

func (m *psRatingRepo) DeleteAll(ctx context.Context) error { return nil }

func (m *psRatingRepo) GetEventDeltaForUser(ctx context.Context, userID, eventID int64) (float64, error) {
	if m.eventDelta == nil {
		return 0, nil
	}
	if byEvent, ok := m.eventDelta[userID]; ok {
		return byEvent[eventID], nil
	}
	return 0, nil
}

type psEventRepo struct {
	events []model.LeagueEvent
	total  int
}

func (m *psEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	return &model.LeagueEvent{EventID: id}, nil
}

func (m *psEventRepo) ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	return nil, nil
}

func (m *psEventRepo) ListDone(ctx context.Context) ([]model.LeagueEvent, error) { return nil, nil }

func (m *psEventRepo) Create(ctx context.Context, e *model.LeagueEvent) (int64, error) {
	return 1, nil
}

func (m *psEventRepo) UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error {
	return nil
}

func (m *psEventRepo) ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error) {
	return m.events, m.total, nil
}

type psGroupRepo struct {
	groups         map[int64][]model.GroupPlayer // groupID → players
	groupsByEvent  map[int64][]model.GroupPlayer // userID,eventID key (not used here — simplified)
	groupDetail    map[int64]*model.Group
	groupMatches   map[int64][]model.Match
}

func newPSGroupRepo() *psGroupRepo {
	return &psGroupRepo{
		groups:       make(map[int64][]model.GroupPlayer),
		groupDetail:  make(map[int64]*model.Group),
		groupMatches: make(map[int64][]model.Match),
	}
}

func (m *psGroupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	if g, ok := m.groupDetail[id]; ok {
		return g, nil
	}
	return &model.Group{GroupID: id}, nil
}

func (m *psGroupRepo) ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error) {
	return nil, nil
}

func (m *psGroupRepo) Create(ctx context.Context, g *model.Group) (int64, error) { return 1, nil }

func (m *psGroupRepo) UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error {
	return nil
}

func (m *psGroupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	return m.groups[groupID], nil
}

func (m *psGroupRepo) GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *psGroupRepo) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	return 1, nil
}

func (m *psGroupRepo) UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error { return nil }

func (m *psGroupRepo) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }

func (m *psGroupRepo) ResetGroupPlayers(ctx context.Context, groupID int64) error { return nil }

func (m *psGroupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *psGroupRepo) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	return nil, nil
}

func (m *psGroupRepo) SetPlayerStatus(ctx context.Context, groupPlayerID int64, status model.PlayerStatus) error {
	return nil
}

type psMatchRepo struct {
	matches map[int64][]model.Match // groupID → matches
}

func (m *psMatchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) { return nil, nil }

func (m *psMatchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	return m.matches[groupID], nil
}

func (m *psMatchRepo) Create(ctx context.Context, match *model.Match) (int64, error) { return 1, nil }

func (m *psMatchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error {
	return nil
}

func (m *psMatchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	return nil
}

func (m *psMatchRepo) BulkCreate(ctx context.Context, matches []model.Match) error { return nil }

func (m *psMatchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error { return nil }

func (m *psMatchRepo) SetWithdraw(ctx context.Context, matchID int64, position int) error {
	return nil
}

func (m *psMatchRepo) SetTableNumber(ctx context.Context, matchID int64, tableNumber int) error {
	return nil
}

func (m *psMatchRepo) ResetScore(ctx context.Context, matchID int64) error { return nil }

func (m *psMatchRepo) ListInProgressByEvent(ctx context.Context, eventID int64) ([]int, error) {
	return nil, nil
}

// --- nopProfileService ---

type nopProfileService struct{}

func (n *nopProfileService) GetProfile(ctx context.Context, userID int64) (*model.PlayerProfileDetail, error) {
	return nil, nil
}

func (n *nopProfileService) UpsertProfile(ctx context.Context, userID int64, req UpsertProfileRequest) (*model.PlayerProfileDetail, error) {
	return nil, nil
}

func (n *nopProfileService) ListCountries(ctx context.Context) ([]model.Country, error) {
	return nil, nil
}

func (n *nopProfileService) ListCities(ctx context.Context, countryID int) ([]model.City, error) {
	return nil, nil
}

func (n *nopProfileService) AddCity(ctx context.Context, name string, countryID int) (*model.City, error) {
	return nil, nil
}

func (n *nopProfileService) ListBlades(ctx context.Context) ([]model.Blade, error) { return nil, nil }

func (n *nopProfileService) AddBlade(ctx context.Context, name string) (*model.Blade, error) {
	return nil, nil
}

func (n *nopProfileService) ListRubbers(ctx context.Context) ([]model.Rubber, error) {
	return nil, nil
}

func (n *nopProfileService) AddRubber(ctx context.Context, name string) (*model.Rubber, error) {
	return nil, nil
}

// --- CreatePlayer tests ---

func TestCreatePlayer_Success(t *testing.T) {
	ur := newPSUserRepo()
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	u, err := svc.CreatePlayer(context.Background(), "Alice", "Smith", "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u == nil {
		t.Fatal("expected non-nil user")
	}
	if u.FirstName != "Alice" {
		t.Errorf("expected FirstName=Alice, got %s", u.FirstName)
	}
	// Default Glicko2 rating.
	if u.CurrentRating != 1500 {
		t.Errorf("expected default rating 1500, got %v", u.CurrentRating)
	}
}

func TestCreatePlayer_RepoError(t *testing.T) {
	ur := newPSUserRepo()
	ur.createErr = errors.New("db error")
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	_, err := svc.CreatePlayer(context.Background(), "Bob", "Jones", "bob@example.com")
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

// --- ImportCSV tests ---

func TestImportCSV_Success(t *testing.T) {
	ur := newPSUserRepo()
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	csv := "first_name,last_name,email\nAlice,Smith,alice@example.com\nBob,Jones,bob@example.com\n"
	result, err := svc.ImportCSV(context.Background(), strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Imported != 2 {
		t.Errorf("expected 2 imported, got %d", result.Imported)
	}
}

func TestImportCSV_SkipsDuplicate(t *testing.T) {
	ur := newPSUserRepo()
	// Pre-seed an existing user.
	ur.byEmail["alice@example.com"] = &model.User{Email: "alice@example.com"}

	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	csv := "first_name,last_name,email\nAlice,Smith,alice@example.com\nBob,Jones,bob@example.com\n"
	result, err := svc.ImportCSV(context.Background(), strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}
	if result.Imported != 1 {
		t.Errorf("expected 1 imported, got %d", result.Imported)
	}
}

func TestImportCSV_MissingRequiredColumn(t *testing.T) {
	ur := newPSUserRepo()
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	// Missing email column.
	csv := "first_name,last_name\nAlice,Smith\n"
	_, err := svc.ImportCSV(context.Background(), strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing column")
	}
}

func TestImportCSV_EmptyRow(t *testing.T) {
	ur := newPSUserRepo()
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	csv := "first_name,last_name,email\n,,\nBob,Jones,bob@example.com\n"
	result, err := svc.ImportCSV(context.Background(), strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 row error, got %d", len(result.Errors))
	}
	if result.Imported != 1 {
		t.Errorf("expected 1 imported, got %d", result.Imported)
	}
}

func TestImportCSV_WithInitialRating(t *testing.T) {
	ur := newPSUserRepo()
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	csv := "first_name,last_name,email,initial_rating\nAlice,Smith,alice@example.com,1800\n"
	result, err := svc.ImportCSV(context.Background(), strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Imported != 1 {
		t.Errorf("expected 1 imported, got %d", result.Imported)
	}
	// Verify the rating was set.
	if ur.users[1].CurrentRating != 1800 {
		t.Errorf("expected CurrentRating=1800, got %v", ur.users[1].CurrentRating)
	}
}

// --- GetProfile tests ---

func TestGetPlayerProfile_Success(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice", LastName: "Smith"}

	rr := &psRatingRepo{history: map[int64][]model.RatingHistory{
		1: {{UserID: 1, MatchID: 10, Delta: 15.0, Rating: 1515.0}},
	}}
	svc := NewPlayerService(ur, rr, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	profile, err := svc.GetProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UserID != 1 {
		t.Errorf("expected userID 1, got %d", profile.UserID)
	}
	if len(profile.RatingHistory) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(profile.RatingHistory))
	}
}

func TestGetPlayerProfile_UserNotFound(t *testing.T) {
	ur := newPSUserRepo()
	rr := &psRatingRepo{history: map[int64][]model.RatingHistory{}}
	svc := NewPlayerService(ur, rr, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	_, err := svc.GetProfile(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

// --- ListPlayers tests ---

func TestListPlayers_All(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice"}
	ur.users[2] = &model.User{UserID: 2, FirstName: "Bob"}

	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	players, err := svc.ListPlayers(context.Background(), "", 10, 0, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 2 {
		t.Errorf("expected 2 players, got %d", len(players))
	}
}

func TestListPlayers_Search(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice"}
	ur.users[2] = &model.User{UserID: 2, FirstName: "Bob"}

	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	players, err := svc.ListPlayers(context.Background(), "Alice", 10, 0, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(players) != 1 {
		t.Errorf("expected 1 player, got %d", len(players))
	}
}

func TestListPlayers_DefaultLimit(t *testing.T) {
	ur := newPSUserRepo()
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, &psEventRepo{}, newPSGroupRepo(), &psMatchRepo{}, &nopProfileService{})

	// limit=0 should default to 50
	_, err := svc.ListPlayers(context.Background(), "", 0, 0, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- GetPlayerEvents tests ---

func TestGetPlayerEvents_Empty(t *testing.T) {
	ur := newPSUserRepo()
	er := &psEventRepo{events: nil, total: 0}
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, er, newPSGroupRepo(), &psMatchRepo{}, nil)

	page, err := svc.GetPlayerEvents(context.Background(), 1, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("expected 0 total, got %d", page.Total)
	}
	if page.Limit != 5 {
		t.Errorf("expected limit 5, got %d", page.Limit)
	}
}

func TestGetPlayerEvents_DefaultLimit(t *testing.T) {
	ur := newPSUserRepo()
	er := &psEventRepo{events: nil, total: 0}
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, er, newPSGroupRepo(), &psMatchRepo{}, nil)

	// limit=0 → default 5
	page, err := svc.GetPlayerEvents(context.Background(), 1, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Limit != 5 {
		t.Errorf("expected default limit 5, got %d", page.Limit)
	}
}

func TestGetPlayerEvents_ExceedLimit(t *testing.T) {
	ur := newPSUserRepo()
	er := &psEventRepo{events: nil, total: 0}
	svc := NewPlayerService(ur, &psRatingRepo{history: map[int64][]model.RatingHistory{}}, er, newPSGroupRepo(), &psMatchRepo{}, nil)

	// limit=100 → clamped to 5
	page, err := svc.GetPlayerEvents(context.Background(), 1, 100, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Limit != 5 {
		t.Errorf("expected clamped limit 5, got %d", page.Limit)
	}
}

func TestGetPlayerEvents_WithEvents(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice"}
	ur.users[2] = &model.User{UserID: 2, FirstName: "Bob"}

	events := []model.LeagueEvent{
		{EventID: 10, LeagueID: 1, Status: model.EventDone},
	}
	er := &psEventRepo{events: events, total: 1}

	// Use a simple group repo that returns no player groups in event (coverage for the events loop).
	groupRepo := newPSGroupRepo()

	rr := &psRatingRepo{history: map[int64][]model.RatingHistory{}}
	svc := NewPlayerService(ur, rr, er, groupRepo, &psMatchRepo{}, nil)

	page, err := svc.GetPlayerEvents(context.Background(), 1, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(page.Events))
	}
}
