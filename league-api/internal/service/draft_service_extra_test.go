package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"league-api/internal/model"
)

// --- Additional mocks for draft service tests ---

type draftMockLeagueRepo struct {
	leagues   map[int64]*model.League
	updateErr error
	assignErr error
}

func (m *draftMockLeagueRepo) GetByID(ctx context.Context, id int64) (*model.League, error) {
	if l, ok := m.leagues[id]; ok {
		return l, nil
	}
	return nil, errors.New("not found")
}

func (m *draftMockLeagueRepo) List(ctx context.Context) ([]model.League, error)   { return nil, nil }

func (m *draftMockLeagueRepo) ListWithStats(ctx context.Context) ([]model.LeagueWithStats, error) {
	return nil, nil
}

func (m *draftMockLeagueRepo) ListMaintainers(ctx context.Context, leagueID int64, limit int) ([]model.LeagueMaintainer, error) {
	return nil, nil
}

func (m *draftMockLeagueRepo) Create(ctx context.Context, l *model.League) (int64, error) {
	return 1, nil
}

func (m *draftMockLeagueRepo) UpdateConfig(ctx context.Context, id int64, config model.LeagueConfig) error {
	return m.updateErr
}

func (m *draftMockLeagueRepo) AssignRole(ctx context.Context, ur model.UserRole) error {
	return m.assignErr
}

func (m *draftMockLeagueRepo) RemoveRole(ctx context.Context, userID, leagueID int64, roleID int) error {
	return nil
}

func (m *draftMockLeagueRepo) GetUserRoles(ctx context.Context, userID, leagueID int64) ([]model.UserRole, error) {
	return nil, nil
}

func (m *draftMockLeagueRepo) GetAllUserRoles(ctx context.Context, userID int64) ([]model.UserRole, error) {
	return nil, nil
}

func (m *draftMockLeagueRepo) ListLeagueRoles(ctx context.Context, leagueID int64) ([]model.UserRole, error) {
	return nil, nil
}

type draftMockMatchRepo struct {
	matches  map[int64][]model.Match // groupID → matches
}

func (m *draftMockMatchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) {
	return nil, nil
}

func (m *draftMockMatchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	return m.matches[groupID], nil
}

func (m *draftMockMatchRepo) Create(ctx context.Context, match *model.Match) (int64, error) {
	return 1, nil
}

func (m *draftMockMatchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error {
	return nil
}

func (m *draftMockMatchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	return nil
}

func (m *draftMockMatchRepo) BulkCreate(ctx context.Context, matches []model.Match) error { return nil }

func (m *draftMockMatchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error { return nil }

func (m *draftMockMatchRepo) SetWithdraw(ctx context.Context, matchID int64, position int) error {
	return nil
}

func (m *draftMockMatchRepo) SetTableNumber(ctx context.Context, matchID int64, tableNumber int) error {
	return nil
}

func (m *draftMockMatchRepo) ResetScore(ctx context.Context, matchID int64) error { return nil }

func (m *draftMockMatchRepo) ListInProgressByEvent(ctx context.Context, eventID int64) ([]int, error) {
	return nil, nil
}

// nopMatchService implements MatchService with no-op logic.
type nopMatchService struct{}

func (s *nopMatchService) UpdateScore(ctx context.Context, matchID int64, score1, score2 int16, gamesToWin int, withdraw1, withdraw2 bool) error {
	return nil
}

func (s *nopMatchService) RecalcGroupPoints(ctx context.Context, groupID int64) error { return nil }

func (s *nopMatchService) SetTableNumber(ctx context.Context, matchID int64, tableNumber int, eventID int64) error {
	return nil
}

func (s *nopMatchService) ResetScore(ctx context.Context, matchID int64) error { return nil }

func (s *nopMatchService) ListInProgressByEvent(ctx context.Context, eventID int64) ([]int, error) {
	return nil, nil
}

// nopRatingService implements RatingService with no-op logic.
type nopRatingService struct {
	calcErr   error
	deleteErr error
}

func (s *nopRatingService) CalculateGroupRatings(ctx context.Context, groupID int64) error {
	return s.calcErr
}

func (s *nopRatingService) RecalculateGroupRatings(ctx context.Context, groupID int64) error {
	return nil
}

func (s *nopRatingService) DeleteGroupRatings(ctx context.Context, groupID int64) error {
	return s.deleteErr
}

func (s *nopRatingService) RecalculateAllRatings(ctx context.Context) (RecalcResult, error) {
	return RecalcResult{}, nil
}

// nopGroupService implements GroupService with no-op logic.
type nopGroupService struct {
	needsManual []int64
	calcErr     error
}

func (s *nopGroupService) GenerateRoundRobin(ctx context.Context, groupID int64) error { return nil }

func (s *nopGroupService) CalculatePlacements(ctx context.Context, groupID int64) ([]int64, error) {
	return s.needsManual, s.calcErr
}

func (s *nopGroupService) SetManualPlace(ctx context.Context, groupPlayerID int64, place int16) error {
	return nil
}

func (s *nopGroupService) AddNonCalculatedPlayer(ctx context.Context, groupID, userID int64) error {
	return nil
}

func (s *nopGroupService) GetGroupDetail(ctx context.Context, groupID int64) (*model.Group, []model.GroupPlayer, []model.Match, error) {
	return &model.Group{}, nil, nil, nil
}

func (s *nopGroupService) ListGroups(ctx context.Context, eventID int64) ([]model.Group, error) {
	return nil, nil
}

func (s *nopGroupService) CreateGroup(ctx context.Context, eventID int64, division string, groupNo int, scheduled time.Time) (*model.Group, error) {
	return nil, nil
}

func (s *nopGroupService) SeedPlayer(ctx context.Context, groupID, userID int64) error { return nil }

func (s *nopGroupService) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }

func (s *nopGroupService) SetPlayerStatus(ctx context.Context, groupID, groupPlayerID int64, status model.PlayerStatus) error {
	return nil
}

func (s *nopGroupService) AddPlayerToActiveGroup(ctx context.Context, groupID, userID int64) error {
	return nil
}

// groupWithPlayers is a helper to build a draftMockGroupRepo with players in specific groups.
type draftGroupRepoWithPlayers struct {
	draftMockGroupRepo
	playersByGroup map[int64][]model.GroupPlayer
}

func (m *draftGroupRepoWithPlayers) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	return m.playersByGroup[groupID], nil
}

func (m *draftGroupRepoWithPlayers) GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error) {
	all := m.playersByGroup[groupID]
	var result []model.GroupPlayer
	for _, p := range all {
		switch moves {
		case model.MoveUp:
			if p.Advances {
				result = append(result, p)
			}
		case model.MoveDown:
			if p.Recedes {
				result = append(result, p)
			}
		case model.MoveStay:
			if !p.Advances && !p.Recedes {
				result = append(result, p)
			}
		}
	}
	return result, nil
}

func (m *draftGroupRepoWithPlayers) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	users := make([]model.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, model.User{UserID: id})
	}
	return users, nil
}

// makeLeague creates a league with config.
func makeLeague(leagueID int64, advances, recedes int) *model.League {
	return &model.League{
		LeagueID: leagueID,
		Config: model.LeagueConfig{
			NumberOfAdvances: advances,
			NumberOfRecedes:  recedes,
			GamesToWin:       3,
		},
	}
}

// makeEventForLeague creates an in-progress event for a league.
func makeEventForLeague(eventID, leagueID int64) *model.LeagueEvent {
	return &model.LeagueEvent{EventID: eventID, LeagueID: leagueID, Status: model.EventInProgress}
}

// --- ReopenGroup tests ---

func TestReopenGroup_Success(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupDone}
	ev := &model.LeagueEvent{EventID: 10, Status: model.EventInProgress}

	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{10: ev},
	}
	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
		ratingSvc: &nopRatingService{},
	}

	err := svc.ReopenGroup(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gr.statusCalls) != 1 || gr.statusCalls[0] != model.GroupInProgress {
		t.Errorf("expected GroupInProgress status call, got %v", gr.statusCalls)
	}
}

func TestReopenGroup_GroupNotDone(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	err := svc.ReopenGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for non-DONE group")
	}
}

func TestReopenGroup_EventNotInProgress(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupDone}
	ev := &model.LeagueEvent{EventID: 10, Status: model.EventDone}

	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{10: ev},
	}
	svc := &draftService{groupRepo: gr, eventRepo: er, ratingSvc: &nopRatingService{}}

	err := svc.ReopenGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when event is not IN_PROGRESS")
	}
}

func TestReopenGroup_DeleteRatingsError(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupDone}
	ev := &model.LeagueEvent{EventID: 10, Status: model.EventInProgress}

	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{10: ev}}
	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
		ratingSvc: &nopRatingService{deleteErr: errors.New("delete error")},
	}

	err := svc.ReopenGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from DeleteGroupRatings")
	}
}

// --- FinishGroup tests ---

func TestFinishGroup_AlreadyDone(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupDone}
	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	err := svc.FinishGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for already DONE group")
	}
}

func TestFinishGroup_MatchNotDone(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
	}
	mr := &draftMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {{MatchID: 1, GroupID: 1, Status: model.MatchDraft}},
		},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}
	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
		matchRepo: mr,
	}

	err := svc.FinishGroup(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for undone match")
	}
}

func TestFinishGroup_RequiresManualPlacements(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
		players:   map[int64][]model.GroupPlayer{},
	}
	mr := &draftMockMatchRepo{
		matches: map[int64][]model.Match{1: {}}, // no matches
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}

	// nopGroupService returns 3 players needing manual placement.
	groupSvc := &nopGroupService{needsManual: []int64{1, 2, 3}}

	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
		matchRepo: mr,
		matchSvc:  &nopMatchService{},
		groupSvc:  groupSvc,
		hub:       nil, // nil hub is ok since FinishGroup checks nil before broadcasting
	}

	err := svc.FinishGroup(context.Background(), 1)
	// Should return nil — manual placement required, not an error.
	if err != nil {
		t.Fatalf("expected nil when manual placement required, got: %v", err)
	}
}

func TestFinishGroup_Success(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	ev := makeEventForLeague(10, 1)
	league := makeLeague(1, 1, 1)

	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 1, UserID: 1, Place: 1},
		{GroupPlayerID: 2, GroupID: 1, UserID: 2, Place: 2},
		{GroupPlayerID: 3, GroupID: 1, UserID: 3, Place: 3},
	}

	gr := &draftGroupRepoWithPlayers{
		draftMockGroupRepo: draftMockGroupRepo{
			groupByID: map[int64]*model.Group{1: grp},
			groups:    map[int64][]model.Group{},
		},
		playersByGroup: map[int64][]model.GroupPlayer{1: players},
	}
	mr := &draftMockMatchRepo{matches: map[int64][]model.Match{1: {}}}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{10: ev}}
	lr := &draftMockLeagueRepo{leagues: map[int64]*model.League{1: league}}

	svc := &draftService{
		leagueRepo: lr,
		groupRepo:  gr,
		eventRepo:  er,
		matchRepo:  mr,
		matchSvc:   &nopMatchService{},
		ratingSvc:  &nopRatingService{},
		groupSvc:   &nopGroupService{needsManual: nil},
		hub:        nil,
	}

	err := svc.FinishGroup(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- SetManualPlacements tests ---

func TestSetManualPlacements_Success(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	ev := makeEventForLeague(10, 1)
	league := makeLeague(1, 1, 1)

	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 1, UserID: 1, Points: 3},
		{GroupPlayerID: 2, GroupID: 1, UserID: 2, Points: 3},
	}

	gr := &draftGroupRepoWithPlayers{
		draftMockGroupRepo: draftMockGroupRepo{
			groupByID: map[int64]*model.Group{1: grp},
			groups:    map[int64][]model.Group{},
		},
		playersByGroup: map[int64][]model.GroupPlayer{1: players},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{10: ev}}
	lr := &draftMockLeagueRepo{leagues: map[int64]*model.League{1: league}}

	svc := &draftService{
		leagueRepo: lr,
		groupRepo:  gr,
		eventRepo:  er,
		ratingSvc:  &nopRatingService{},
		hub:        nil,
	}

	err := svc.SetManualPlacements(context.Background(), 1, []int64{2, 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetManualPlacements_InvalidPlayerID(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 1, UserID: 1, Points: 3},
	}

	gr := &draftGroupRepoWithPlayers{
		draftMockGroupRepo: draftMockGroupRepo{
			groupByID: map[int64]*model.Group{1: grp},
			groups:    map[int64][]model.Group{},
		},
		playersByGroup: map[int64][]model.GroupPlayer{1: players},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}

	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
	}

	// Player ID 99 doesn't exist.
	err := svc.SetManualPlacements(context.Background(), 1, []int64{99})
	if err == nil {
		t.Fatal("expected error for invalid groupPlayerID")
	}
}

// --- CreateDraft tests ---

func TestCreateDraft_GroupNotDone(t *testing.T) {
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {inProgressGroup(10, 1)},
		},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}

	svc := &draftService{groupRepo: gr, eventRepo: er}
	_, err := svc.CreateDraft(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected error for non-DONE group")
	}
}

func TestCreateDraft_AlreadyActiveEvent(t *testing.T) {
	existingEvent := &model.LeagueEvent{EventID: 5, LeagueID: 1, Status: model.EventDraft}
	finishedEvent := &model.LeagueEvent{EventID: 1, LeagueID: 1, Status: model.EventDone}

	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{1: {doneGroup(10, 1)}},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{1: finishedEvent, 5: existingEvent},
	}
	er.events[1] = finishedEvent

	// Inject a custom ListByLeague that returns an active event.
	type customEventRepo struct {
		draftMockEventRepo
		activeEvents []model.LeagueEvent
	}

	cer := &struct {
		draftMockEventRepo
		activeEvents []model.LeagueEvent
	}{
		draftMockEventRepo: draftMockEventRepo{
			events: map[int64]*model.LeagueEvent{1: finishedEvent},
		},
		activeEvents: []model.LeagueEvent{*existingEvent},
	}
	_ = cer

	// Use a simpler approach — build a mock that overrides ListByLeague directly.
	type eventRepoWithActiveLeague struct {
		base         *draftMockEventRepo
		leagueEvents map[int64][]model.LeagueEvent
	}

	fullER := &fullEventRepo{
		base: er,
		leagueEvents: map[int64][]model.LeagueEvent{
			1: {*existingEvent},
		},
	}

	svc := &draftService{groupRepo: gr, eventRepo: fullER}
	_, err := svc.CreateDraft(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected error for existing active event")
	}
}

// fullEventRepo extends draftMockEventRepo with per-league listing.
type fullEventRepo struct {
	base         *draftMockEventRepo
	leagueEvents map[int64][]model.LeagueEvent
}

func (m *fullEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	return m.base.GetByID(ctx, id)
}

func (m *fullEventRepo) ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	return m.leagueEvents[leagueID], nil
}

func (m *fullEventRepo) ListDone(ctx context.Context) ([]model.LeagueEvent, error) { return nil, nil }

func (m *fullEventRepo) Create(ctx context.Context, e *model.LeagueEvent) (int64, error) {
	return m.base.Create(ctx, e)
}

func (m *fullEventRepo) UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error {
	return m.base.UpdateStatus(ctx, id, status)
}

func (m *fullEventRepo) ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error) {
	return nil, 0, nil
}

func TestCreateDraft_Success_SingleGroup(t *testing.T) {
	finishedEvent := &model.LeagueEvent{EventID: 1, LeagueID: 1, Status: model.EventDone}

	// One done group with players: 1 advances, 1 stays, 1 recedes.
	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 10, UserID: 1, Place: 1, Advances: true, Recedes: false},
		{GroupPlayerID: 2, GroupID: 10, UserID: 2, Place: 2, Advances: false, Recedes: false},
		{GroupPlayerID: 3, GroupID: 10, UserID: 3, Place: 3, Advances: false, Recedes: true},
	}

	gr := &draftGroupRepoWithPlayers{
		draftMockGroupRepo: draftMockGroupRepo{
			groups:    map[int64][]model.Group{1: {doneGroup(10, 1)}},
			groupByID: map[int64]*model.Group{10: {GroupID: 10, EventID: 1, Status: model.GroupDone}},
		},
		playersByGroup: map[int64][]model.GroupPlayer{10: players},
	}

	er := &fullEventRepo{
		base: &draftMockEventRepo{
			events:   map[int64]*model.LeagueEvent{1: finishedEvent, 42: {EventID: 42, LeagueID: 1, Status: model.EventDraft}},
			createID: 42,
		},
		leagueEvents: map[int64][]model.LeagueEvent{1: {}}, // no active events
	}

	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
	}

	result, err := svc.CreateDraft(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil event returned")
	}
}

// --- RecreateDraft tests ---

func TestRecreateDraft_NotDraftEvent(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 1, LeagueID: 1, Status: model.EventInProgress}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{1: ev}}
	gr := &draftMockGroupRepo{groups: map[int64][]model.Group{}}
	lr := &draftMockLeagueRepo{leagues: map[int64]*model.League{}}

	svc := &draftService{leagueRepo: lr, groupRepo: gr, eventRepo: er}
	err := svc.RecreateDraft(context.Background(), 1, model.LeagueConfig{})
	if err == nil {
		t.Fatal("expected error for non-DRAFT event")
	}
}

func TestRecreateDraft_DraftEvent(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 1, LeagueID: 1, Status: model.EventDraft}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{1: ev}}
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{1: {}},
	}
	lr := &draftMockLeagueRepo{leagues: map[int64]*model.League{1: {LeagueID: 1}}}

	svc := &draftService{leagueRepo: lr, groupRepo: gr, eventRepo: er}
	err := svc.RecreateDraft(context.Background(), 1, model.LeagueConfig{GamesToWin: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- orderedDivisions tests ---

func TestOrderedDivisions(t *testing.T) {
	divGroups := map[string][]model.Group{
		"B":           {{GroupID: 1}},
		"Superleague": {{GroupID: 2}},
		"A":           {{GroupID: 3}},
		"C":           {{GroupID: 4}},
	}

	ordered := orderedDivisions(divGroups)

	// Verify order: Superleague, A, B, C.
	expected := []string{"Superleague", "A", "B", "C"}
	if len(ordered) != len(expected) {
		t.Fatalf("expected %d divisions, got %d", len(expected), len(ordered))
	}
	for i, div := range expected {
		if ordered[i] != div {
			t.Errorf("position %d: expected %s, got %s", i, div, ordered[i])
		}
	}
}
