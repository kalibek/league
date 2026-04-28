package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"league-api/internal/model"
)

// --- event service mocks ---

type evtMockEventRepo struct {
	events      map[int64]*model.LeagueEvent
	byLeague    map[int64][]model.LeagueEvent
	createID    int64
	createErr   error
	updateErr   error
	statusCalls []model.EventStatus
}

func (m *evtMockEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	if e, ok := m.events[id]; ok {
		return e, nil
	}
	return nil, errors.New("not found")
}

func (m *evtMockEventRepo) ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	return m.byLeague[leagueID], nil
}

func (m *evtMockEventRepo) ListDone(ctx context.Context) ([]model.LeagueEvent, error) { return nil, nil }

func (m *evtMockEventRepo) Create(ctx context.Context, e *model.LeagueEvent) (int64, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 42
	}
	m.events[m.createID] = e
	return m.createID, nil
}

func (m *evtMockEventRepo) UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error {
	m.statusCalls = append(m.statusCalls, status)
	if m.updateErr != nil {
		return m.updateErr
	}
	if e, ok := m.events[id]; ok {
		e.Status = status
	}
	return nil
}

func (m *evtMockEventRepo) ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error) {
	return nil, 0, nil
}

type evtMockGroupRepo struct {
	groups  map[int64][]model.Group
	players map[int64][]model.GroupPlayer
	created []model.Group
	statusCalls map[int64]model.GroupStatus
}

func (m *evtMockGroupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	return &model.Group{GroupID: id, EventID: 1}, nil
}

func (m *evtMockGroupRepo) ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error) {
	return m.groups[eventID], nil
}

func (m *evtMockGroupRepo) Create(ctx context.Context, g *model.Group) (int64, error) {
	m.created = append(m.created, *g)
	return int64(len(m.created)), nil
}

func (m *evtMockGroupRepo) UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error {
	if m.statusCalls == nil {
		m.statusCalls = make(map[int64]model.GroupStatus)
	}
	m.statusCalls[id] = status
	return nil
}

func (m *evtMockGroupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	return m.players[groupID], nil
}

func (m *evtMockGroupRepo) GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *evtMockGroupRepo) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	return 1, nil
}

func (m *evtMockGroupRepo) UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error { return nil }

func (m *evtMockGroupRepo) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }

func (m *evtMockGroupRepo) ResetGroupPlayers(ctx context.Context, groupID int64) error { return nil }

func (m *evtMockGroupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *evtMockGroupRepo) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	return nil, nil
}

type evtMockMatchRepo struct {
	bulkCalls int
}

func (m *evtMockMatchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) { return nil, nil }

func (m *evtMockMatchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	return nil, nil
}

func (m *evtMockMatchRepo) Create(ctx context.Context, match *model.Match) (int64, error) { return 1, nil }

func (m *evtMockMatchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error {
	return nil
}

func (m *evtMockMatchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	return nil
}

func (m *evtMockMatchRepo) BulkCreate(ctx context.Context, matches []model.Match) error {
	m.bulkCalls++
	return nil
}

func (m *evtMockMatchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error { return nil }

func (m *evtMockMatchRepo) SetWithdraw(ctx context.Context, matchID int64, position int) error {
	return nil
}

type evtMockUserRepo struct{}

func (m *evtMockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return &model.User{UserID: id}, nil
}

func (m *evtMockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}

func (m *evtMockUserRepo) Create(ctx context.Context, u *model.User) (int64, error) { return 1, nil }

func (m *evtMockUserRepo) List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *evtMockUserRepo) Search(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *evtMockUserRepo) UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error {
	return nil
}

func (m *evtMockUserRepo) ResetAllRatings(ctx context.Context) error { return nil }

func (m *evtMockUserRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	return nil
}

func (m *evtMockUserRepo) UpdateName(ctx context.Context, userID int64, firstName, lastName string) error {
	return nil
}

// --- CreateDraftEvent tests ---

func TestCreateDraftEvent_Success(t *testing.T) {
	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{},
		byLeague: map[int64][]model.LeagueEvent{1: {}},
		createID: 10,
	}
	gr := &evtMockGroupRepo{groups: map[int64][]model.Group{}}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)

	ev, err := svc.CreateDraftEvent(context.Background(), 1, "Test Event", start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev == nil {
		t.Fatal("expected event to be returned, got nil")
	}
}

func TestCreateDraftEvent_AlreadyActiveEvent(t *testing.T) {
	existing := model.LeagueEvent{EventID: 1, LeagueID: 1, Status: model.EventDraft}
	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{1: &existing},
		byLeague: map[int64][]model.LeagueEvent{1: {existing}},
	}
	gr := &evtMockGroupRepo{}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	_, err := svc.CreateDraftEvent(context.Background(), 1, "Test Event", time.Now(), time.Now().Add(24*time.Hour))
	if err == nil {
		t.Fatal("expected error for duplicate active event")
	}
}

func TestCreateDraftEvent_InProgressBlocks(t *testing.T) {
	existing := model.LeagueEvent{EventID: 2, LeagueID: 1, Status: model.EventInProgress}
	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{2: &existing},
		byLeague: map[int64][]model.LeagueEvent{1: {existing}},
	}
	gr := &evtMockGroupRepo{}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	_, err := svc.CreateDraftEvent(context.Background(), 1, "Test Event", time.Now(), time.Now().Add(24*time.Hour))
	if err == nil {
		t.Fatal("expected error for in-progress event")
	}
}

// --- StartEvent tests ---

func TestStartEvent_Success(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 5, Status: model.EventDraft}
	gp1, gp2 := int64(1), int64(2)
	groups := []model.Group{
		{GroupID: 10, EventID: 5, Status: model.GroupDraft},
	}
	players := []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 10, UserID: 1},
		{GroupPlayerID: gp2, GroupID: 10, UserID: 2},
	}

	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{5: ev},
		byLeague: map[int64][]model.LeagueEvent{},
	}
	gr := &evtMockGroupRepo{
		groups:  map[int64][]model.Group{5: groups},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	err := svc.StartEvent(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have created matches and updated status
	if mr.bulkCalls != 1 {
		t.Errorf("expected 1 BulkCreate call, got %d", mr.bulkCalls)
	}
	if len(er.statusCalls) != 1 || er.statusCalls[0] != model.EventInProgress {
		t.Errorf("expected EventInProgress status update, got %v", er.statusCalls)
	}
}

func TestStartEvent_NotDraft(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 5, Status: model.EventInProgress}
	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{5: ev},
		byLeague: map[int64][]model.LeagueEvent{},
	}
	gr := &evtMockGroupRepo{groups: map[int64][]model.Group{}}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	err := svc.StartEvent(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error for non-DRAFT event")
	}
}

func TestStartEvent_NoPlayers_NoMatches(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 5, Status: model.EventDraft}
	groups := []model.Group{
		{GroupID: 10, EventID: 5, Status: model.GroupDraft},
	}

	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{5: ev},
		byLeague: map[int64][]model.LeagueEvent{},
	}
	gr := &evtMockGroupRepo{
		groups:  map[int64][]model.Group{5: groups},
		players: map[int64][]model.GroupPlayer{10: {}}, // empty
	}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	err := svc.StartEvent(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No players → no match stubs
	if mr.bulkCalls != 0 {
		t.Errorf("expected 0 BulkCreate calls for empty group, got %d", mr.bulkCalls)
	}
}

func TestStartEvent_NonCalculatedExcluded(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 5, Status: model.EventDraft}
	groups := []model.Group{
		{GroupID: 10, EventID: 5, Status: model.GroupDraft},
	}
	gp1, gp2, gp3 := int64(1), int64(2), int64(3)
	players := []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 10, UserID: 1},
		{GroupPlayerID: gp2, GroupID: 10, UserID: 2},
		{GroupPlayerID: gp3, GroupID: 10, UserID: 3, IsNonCalculated: true},
	}

	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{5: ev},
		byLeague: map[int64][]model.LeagueEvent{},
	}
	gr := &evtMockGroupRepo{
		groups:  map[int64][]model.Group{5: groups},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	err := svc.StartEvent(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only 2 calculated players → 1 match stub
	if mr.bulkCalls != 1 {
		t.Errorf("expected 1 BulkCreate call, got %d", mr.bulkCalls)
	}
}

// --- ListEvents / GetEvent / GetEventDetail tests ---

func TestListEvents(t *testing.T) {
	evts := []model.LeagueEvent{
		{EventID: 1, LeagueID: 10, Status: model.EventDraft},
		{EventID: 2, LeagueID: 10, Status: model.EventDone},
	}
	er := &evtMockEventRepo{
		events:   map[int64]*model.LeagueEvent{},
		byLeague: map[int64][]model.LeagueEvent{10: evts},
	}
	svc := NewEventService(er, &evtMockGroupRepo{}, &evtMockMatchRepo{}, &evtMockUserRepo{})

	list, err := svc.ListEvents(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 events, got %d", len(list))
	}
}

func TestGetEvent_Found(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 7, Status: model.EventDraft}
	er := &evtMockEventRepo{events: map[int64]*model.LeagueEvent{7: ev}}
	svc := NewEventService(er, &evtMockGroupRepo{}, &evtMockMatchRepo{}, &evtMockUserRepo{})

	got, err := svc.GetEvent(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.EventID != 7 {
		t.Errorf("expected eventID 7, got %d", got.EventID)
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	er := &evtMockEventRepo{events: map[int64]*model.LeagueEvent{}}
	svc := NewEventService(er, &evtMockGroupRepo{}, &evtMockMatchRepo{}, &evtMockUserRepo{})

	_, err := svc.GetEvent(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error for missing event")
	}
}

func TestGetEventDetail_WithGroups(t *testing.T) {
	ev := &model.LeagueEvent{EventID: 5, Status: model.EventInProgress}
	groups := []model.Group{
		{GroupID: 10, EventID: 5},
		{GroupID: 11, EventID: 5},
	}
	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 10, UserID: 1},
	}

	er := &evtMockEventRepo{events: map[int64]*model.LeagueEvent{5: ev}}
	gr := &evtMockGroupRepo{
		groups:  map[int64][]model.Group{5: groups},
		players: map[int64][]model.GroupPlayer{10: players, 11: {}},
	}
	mr := &evtMockMatchRepo{}
	ur := &evtMockUserRepo{}

	svc := NewEventService(er, gr, mr, ur)

	detail, err := svc.GetEventDetail(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.EventID != 5 {
		t.Errorf("expected eventID 5, got %d", detail.EventID)
	}
	if len(detail.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(detail.Groups))
	}
}
