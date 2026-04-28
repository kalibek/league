package service

import (
	"context"
	"errors"
	"testing"

	"league-api/internal/model"
)

// --- league service mocks ---

type mockLeagueRepo struct {
	leagues     map[int64]*model.League
	stats       []model.LeagueWithStats
	maintainers map[int64][]model.LeagueMaintainer
	roles       map[int64][]model.UserRole       // leagueID → roles
	userRoles   map[int64]map[int64][]model.UserRole // userID → leagueID → roles
	createErr   error
	updateErr   error
	assignErr   error
	removeErr   error
	nextID      int64
}

func newMockLeagueRepo() *mockLeagueRepo {
	return &mockLeagueRepo{
		leagues:     make(map[int64]*model.League),
		maintainers: make(map[int64][]model.LeagueMaintainer),
		roles:       make(map[int64][]model.UserRole),
		userRoles:   make(map[int64]map[int64][]model.UserRole),
		nextID:      1,
	}
}

func (m *mockLeagueRepo) GetByID(ctx context.Context, id int64) (*model.League, error) {
	if l, ok := m.leagues[id]; ok {
		return l, nil
	}
	return nil, errors.New("not found")
}

func (m *mockLeagueRepo) List(ctx context.Context) ([]model.League, error) {
	result := make([]model.League, 0, len(m.leagues))
	for _, l := range m.leagues {
		result = append(result, *l)
	}
	return result, nil
}

func (m *mockLeagueRepo) ListWithStats(ctx context.Context) ([]model.LeagueWithStats, error) {
	return m.stats, nil
}

func (m *mockLeagueRepo) ListMaintainers(ctx context.Context, leagueID int64, limit int) ([]model.LeagueMaintainer, error) {
	return m.maintainers[leagueID], nil
}

func (m *mockLeagueRepo) Create(ctx context.Context, l *model.League) (int64, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	id := m.nextID
	m.nextID++
	m.leagues[id] = l
	return id, nil
}

func (m *mockLeagueRepo) UpdateConfig(ctx context.Context, id int64, config model.LeagueConfig) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if l, ok := m.leagues[id]; ok {
		l.Config = config
	}
	return nil
}

func (m *mockLeagueRepo) AssignRole(ctx context.Context, ur model.UserRole) error {
	if m.assignErr != nil {
		return m.assignErr
	}
	m.roles[ur.LeagueID] = append(m.roles[ur.LeagueID], ur)
	if m.userRoles[ur.UserID] == nil {
		m.userRoles[ur.UserID] = make(map[int64][]model.UserRole)
	}
	m.userRoles[ur.UserID][ur.LeagueID] = append(m.userRoles[ur.UserID][ur.LeagueID], ur)
	return nil
}

func (m *mockLeagueRepo) RemoveRole(ctx context.Context, userID, leagueID int64, roleID int) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	return nil
}

func (m *mockLeagueRepo) GetUserRoles(ctx context.Context, userID, leagueID int64) ([]model.UserRole, error) {
	if m.userRoles[userID] == nil {
		return nil, nil
	}
	return m.userRoles[userID][leagueID], nil
}

func (m *mockLeagueRepo) GetAllUserRoles(ctx context.Context, userID int64) ([]model.UserRole, error) {
	if m.userRoles[userID] == nil {
		return nil, nil
	}
	var all []model.UserRole
	for _, roles := range m.userRoles[userID] {
		all = append(all, roles...)
	}
	return all, nil
}

func (m *mockLeagueRepo) ListLeagueRoles(ctx context.Context, leagueID int64) ([]model.UserRole, error) {
	return m.roles[leagueID], nil
}

type mockUserRepoForLeague struct {
	users map[int64]*model.User
}

func (m *mockUserRepoForLeague) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return &model.User{UserID: id}, nil
}

func (m *mockUserRepoForLeague) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepoForLeague) Create(ctx context.Context, u *model.User) (int64, error) { return 1, nil }

func (m *mockUserRepoForLeague) List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *mockUserRepoForLeague) Search(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *mockUserRepoForLeague) UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error {
	return nil
}

func (m *mockUserRepoForLeague) ResetAllRatings(ctx context.Context) error { return nil }

func (m *mockUserRepoForLeague) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	return nil
}

func (m *mockUserRepoForLeague) UpdateName(ctx context.Context, userID int64, firstName, lastName string) error {
	return nil
}

// --- Tests ---

func TestCreateLeague_Success(t *testing.T) {
	lr := newMockLeagueRepo()
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	league, err := svc.CreateLeague(context.Background(), 1, "Test League", "desc", model.LeagueConfig{GamesToWin: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if league == nil {
		t.Fatal("expected non-nil league")
	}
	// Creator should have maintainer and player roles assigned.
	roles := lr.roles[1] // leagueID 1
	if len(roles) != 2 {
		t.Errorf("expected 2 roles assigned, got %d", len(roles))
	}
}

func TestCreateLeague_CreateError(t *testing.T) {
	lr := newMockLeagueRepo()
	lr.createErr = errors.New("db error")
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	_, err := svc.CreateLeague(context.Background(), 1, "Test League", "desc", model.LeagueConfig{})
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

func TestCreateLeague_AssignRoleError(t *testing.T) {
	lr := newMockLeagueRepo()
	lr.assignErr = errors.New("assign error")
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	_, err := svc.CreateLeague(context.Background(), 1, "Test League", "desc", model.LeagueConfig{})
	if err == nil {
		t.Fatal("expected error from AssignRole")
	}
}

func TestGetLeague(t *testing.T) {
	lr := newMockLeagueRepo()
	lr.leagues[5] = &model.League{LeagueID: 5, Title: "My League"}
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	league, err := svc.GetLeague(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if league.Title != "My League" {
		t.Errorf("expected 'My League', got %s", league.Title)
	}
}

func TestListLeagues(t *testing.T) {
	lr := newMockLeagueRepo()
	lr.leagues[1] = &model.League{LeagueID: 1}
	lr.leagues[2] = &model.League{LeagueID: 2}
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	leagues, err := svc.ListLeagues(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leagues) != 2 {
		t.Errorf("expected 2 leagues, got %d", len(leagues))
	}
}

func TestListLeagueSummaries(t *testing.T) {
	lr := newMockLeagueRepo()
	lr.stats = []model.LeagueWithStats{
		{LeagueID: 1, Title: "League A"},
		{LeagueID: 2, Title: "League B"},
	}
	lr.maintainers[1] = []model.LeagueMaintainer{{UserID: 1, FirstName: "Alice"}}
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	summaries, err := svc.ListLeagueSummaries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(summaries))
	}
	if len(summaries[0].Maintainers) != 1 {
		t.Errorf("expected 1 maintainer for league 1, got %d", len(summaries[0].Maintainers))
	}
}

func TestUpdateConfig(t *testing.T) {
	lr := newMockLeagueRepo()
	lr.leagues[1] = &model.League{LeagueID: 1}
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	cfg := model.LeagueConfig{GamesToWin: 5, NumberOfAdvances: 2}
	err := svc.UpdateConfig(context.Background(), 1, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lr.leagues[1].Config.GamesToWin != 5 {
		t.Errorf("expected GamesToWin=5, got %d", lr.leagues[1].Config.GamesToWin)
	}
}

func TestAssignRole_ValidRoles(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
	}{
		{"player", "player"},
		{"umpire", "umpire"},
		{"maintainer", "maintainer"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lr := newMockLeagueRepo()
			ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
			svc := NewLeagueService(lr, ur)
			err := svc.AssignRole(context.Background(), 1, 42, tc.roleName)
			if err != nil {
				t.Fatalf("unexpected error for role %s: %v", tc.roleName, err)
			}
		})
	}
}

func TestAssignRole_UnknownRole(t *testing.T) {
	lr := newMockLeagueRepo()
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	err := svc.AssignRole(context.Background(), 1, 42, "superuser")
	if err == nil {
		t.Fatal("expected error for unknown role")
	}
}

func TestRemoveRole_ValidRole(t *testing.T) {
	lr := newMockLeagueRepo()
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	err := svc.RemoveRole(context.Background(), 1, 42, "player")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveRole_UnknownRole(t *testing.T) {
	lr := newMockLeagueRepo()
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	err := svc.RemoveRole(context.Background(), 1, 42, "unknown")
	if err == nil {
		t.Fatal("expected error for unknown role")
	}
}

func TestIsMaintainer_True(t *testing.T) {
	lr := newMockLeagueRepo()
	// Assign maintainer role (roleID=3) directly.
	lr.userRoles[42] = map[int64][]model.UserRole{
		1: {{UserID: 42, LeagueID: 1, RoleID: 3}},
	}
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	if !svc.IsMaintainer(context.Background(), 1, 42) {
		t.Error("expected user 42 to be a maintainer of league 1")
	}
}

func TestIsMaintainer_False(t *testing.T) {
	lr := newMockLeagueRepo()
	ur := &mockUserRepoForLeague{users: map[int64]*model.User{}}
	svc := NewLeagueService(lr, ur)

	if svc.IsMaintainer(context.Background(), 1, 99) {
		t.Error("expected user 99 to not be a maintainer")
	}
}

func TestListLeagueRoles(t *testing.T) {
	lr := newMockLeagueRepo()
	lr.roles[1] = []model.UserRole{
		{UserID: 1, LeagueID: 1, RoleID: 1},
		{UserID: 2, LeagueID: 1, RoleID: 3},
	}
	ur := &mockUserRepoForLeague{
		users: map[int64]*model.User{
			1: {UserID: 1, FirstName: "Alice", LastName: "A", Email: "a@a.com"},
			2: {UserID: 2, FirstName: "Bob", LastName: "B", Email: "b@b.com"},
		},
	}
	svc := NewLeagueService(lr, ur)

	roles, err := svc.ListLeagueRoles(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestRoleNameToID_AllValid(t *testing.T) {
	tests := []struct {
		name     string
		roleName string
		wantID   int
	}{
		{"player", "player", 1},
		{"umpire", "umpire", 2},
		{"maintainer", "maintainer", 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := roleNameToID(tc.roleName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tc.wantID {
				t.Errorf("expected roleID %d for %s, got %d", tc.wantID, tc.roleName, id)
			}
		})
	}
}

func TestRoleNameToID_Unknown(t *testing.T) {
	_, err := roleNameToID("admin")
	if err == nil {
		t.Fatal("expected error for unknown role name")
	}
}

func TestRoleIDToName_AllValid(t *testing.T) {
	tests := []struct {
		id   int
		want string
	}{
		{1, "player"},
		{2, "umpire"},
		{3, "maintainer"},
		{99, "unknown"},
	}
	for _, tc := range tests {
		got := roleIDToName(tc.id)
		if got != tc.want {
			t.Errorf("roleIDToName(%d) = %s, want %s", tc.id, got, tc.want)
		}
	}
}

func TestDivisionRank(t *testing.T) {
	tests := []struct {
		div  string
		want int
	}{
		{"Superleague", 0},
		{"A", 1},
		{"B", 2},
		{"C", 3},
		{"D", 10},
	}
	for _, tc := range tests {
		got := divisionRank(tc.div)
		if got != tc.want {
			t.Errorf("divisionRank(%s) = %d, want %d", tc.div, got, tc.want)
		}
	}
}
