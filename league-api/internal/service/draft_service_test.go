package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"league-api/internal/model"
)

// --- configurable mocks for draftService tests ---

type draftMockGroupRepo struct {
	groups      map[int64][]model.Group  // eventID → groups
	groupByID   map[int64]*model.Group
	statusCalls []model.GroupStatus
	players     map[int64][]model.GroupPlayer // groupID → players
}

func (m *draftMockGroupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	if g, ok := m.groupByID[id]; ok {
		return g, nil
	}
	return nil, errors.New("not found")
}

func (m *draftMockGroupRepo) ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error) {
	return m.groups[eventID], nil
}

func (m *draftMockGroupRepo) Create(ctx context.Context, g *model.Group) (int64, error) { return 1, nil }

func (m *draftMockGroupRepo) UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error {
	m.statusCalls = append(m.statusCalls, status)
	return nil
}

func (m *draftMockGroupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	return m.players[groupID], nil
}

func (m *draftMockGroupRepo) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	return 1, nil
}

func (m *draftMockGroupRepo) UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error { return nil }

func (m *draftMockGroupRepo) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }

func (m *draftMockGroupRepo) ResetGroupPlayers(ctx context.Context, groupID int64) error { return nil }

func (m *draftMockGroupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *draftMockGroupRepo) GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error) {
	all := m.players[groupID]
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

func (m *draftMockGroupRepo) SetPlayerStatus(ctx context.Context, groupPlayerID int64, status model.PlayerStatus) error {
	return nil
}

func (m *draftMockGroupRepo) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	users := make([]model.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, model.User{UserID: id})
	}
	return users, nil
}

type draftMockEventRepo struct {
	events      map[int64]*model.LeagueEvent
	statusCalls []model.EventStatus
	updateErr   error
	createID    int64
}

func (m *draftMockEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	if e, ok := m.events[id]; ok {
		return e, nil
	}
	return nil, errors.New("not found")
}

func (m *draftMockEventRepo) ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	return nil, nil
}

func (m *draftMockEventRepo) Create(ctx context.Context, e *model.LeagueEvent) (int64, error) {
	if m.createID != 0 {
		return m.createID, nil
	}
	return 1, nil
}

func (m *draftMockEventRepo) UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error {
	m.statusCalls = append(m.statusCalls, status)
	return m.updateErr
}

func (m *draftMockEventRepo) ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error) {
	return nil, 0, nil
}

func (m *draftMockEventRepo) ListDone(ctx context.Context) ([]model.LeagueEvent, error) {
	return nil, nil
}

// --- helpers ---

func doneGroup(groupID, eventID int64) model.Group {
	return model.Group{GroupID: groupID, EventID: eventID, Status: model.GroupDone}
}

func inProgressGroup(groupID, eventID int64) model.Group {
	return model.Group{GroupID: groupID, EventID: eventID, Status: model.GroupInProgress}
}

func inProgressEvent(eventID int64) *model.LeagueEvent {
	return &model.LeagueEvent{EventID: eventID, Status: model.EventInProgress}
}

func draftEvent(eventID int64) *model.LeagueEvent {
	return &model.LeagueEvent{EventID: eventID, Status: model.EventDraft}
}

func doneEvent(eventID int64) *model.LeagueEvent {
	return &model.LeagueEvent{EventID: eventID, Status: model.EventDone}
}

// --- FinishEvent tests ---

func TestFinishEvent_Success(t *testing.T) {
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {doneGroup(10, 1), doneGroup(11, 1), doneGroup(12, 1)},
		},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{1: inProgressEvent(1)},
	}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	if err := svc.FinishEvent(context.Background(), 1); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(er.statusCalls) != 1 || er.statusCalls[0] != model.EventDone {
		t.Errorf("expected EventDone status update, got: %v", er.statusCalls)
	}
}

func TestFinishEvent_GroupNotDone(t *testing.T) {
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {doneGroup(10, 1), inProgressGroup(11, 1)},
		},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{1: inProgressEvent(1)},
	}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	err := svc.FinishEvent(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for non-DONE group, got nil")
	}
	if !strings.Contains(err.Error(), "not DONE") {
		t.Errorf("unexpected error message: %v", err)
	}
	if len(er.statusCalls) != 0 {
		t.Error("UpdateStatus should not have been called")
	}
}

func TestFinishEvent_EventNotInProgress(t *testing.T) {
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {doneGroup(10, 1)},
		},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{1: doneEvent(1)},
	}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	err := svc.FinishEvent(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for non-IN_PROGRESS event, got nil")
	}
	if !strings.Contains(err.Error(), "not IN_PROGRESS") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFinishEvent_EventInDraftStatus(t *testing.T) {
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {doneGroup(10, 1)},
		},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{1: draftEvent(1)},
	}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	err := svc.FinishEvent(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for DRAFT event, got nil")
	}
}

func TestFinishEvent_NoGroups(t *testing.T) {
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {},
		},
	}
	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{1: inProgressEvent(1)},
	}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	err := svc.FinishEvent(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for event with no groups, got nil")
	}
	if !strings.Contains(err.Error(), "no groups") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFinishEvent_UpdateStatusError(t *testing.T) {
	gr := &draftMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {doneGroup(10, 1)},
		},
	}
	er := &draftMockEventRepo{
		events:    map[int64]*model.LeagueEvent{1: inProgressEvent(1)},
		updateErr: errors.New("db error"),
	}
	svc := &draftService{groupRepo: gr, eventRepo: er}

	err := svc.FinishEvent(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from UpdateStatus, got nil")
	}
}

// --- FinishGroup DNS tests ---

type draftMockMatchSvc struct {
	recalcCalled bool
}

func (m *draftMockMatchSvc) UpdateScore(ctx context.Context, matchID int64, score1, score2 int16, gamesToWin int, withdraw1, withdraw2 bool) error {
	return nil
}
func (m *draftMockMatchSvc) RecalcGroupPoints(ctx context.Context, groupID int64) error {
	m.recalcCalled = true
	return nil
}
func (m *draftMockMatchSvc) SetTableNumber(ctx context.Context, matchID int64, tableNumber int, eventID int64) error {
	return nil
}
func (m *draftMockMatchSvc) ResetScore(ctx context.Context, matchID int64) error { return nil }
func (m *draftMockMatchSvc) ListInProgressByEvent(ctx context.Context, eventID int64) ([]int, error) {
	return nil, nil
}

type draftMockGroupSvc struct {
	needsManual []int64
}

func (m *draftMockGroupSvc) GenerateRoundRobin(ctx context.Context, groupID int64) error { return nil }
func (m *draftMockGroupSvc) CalculatePlacements(ctx context.Context, groupID int64) ([]int64, error) {
	return m.needsManual, nil
}
func (m *draftMockGroupSvc) SetManualPlace(ctx context.Context, groupPlayerID int64, place int16) error {
	return nil
}
func (m *draftMockGroupSvc) AddNonCalculatedPlayer(ctx context.Context, groupID, userID int64) error {
	return nil
}
func (m *draftMockGroupSvc) GetGroupDetail(ctx context.Context, groupID int64) (*model.Group, []model.GroupPlayer, []model.Match, error) {
	return nil, nil, nil, nil
}
func (m *draftMockGroupSvc) ListGroups(ctx context.Context, eventID int64) ([]model.Group, error) {
	return nil, nil
}
func (m *draftMockGroupSvc) CreateGroup(ctx context.Context, eventID int64, division string, groupNo int, scheduled time.Time) (*model.Group, error) {
	return nil, nil
}
func (m *draftMockGroupSvc) SeedPlayer(ctx context.Context, groupID, userID int64) error { return nil }
func (m *draftMockGroupSvc) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }
func (m *draftMockGroupSvc) SetPlayerStatus(ctx context.Context, groupID, groupPlayerID int64, status model.PlayerStatus) error {
	return nil
}
func (m *draftMockGroupSvc) AddPlayerToActiveGroup(ctx context.Context, groupID, userID int64) error {
	return nil
}

type draftMockRatingSvc struct{}

func (m *draftMockRatingSvc) CalculateGroupRatings(ctx context.Context, groupID int64) error {
	return nil
}
func (m *draftMockRatingSvc) RecalculateGroupRatings(ctx context.Context, groupID int64) error {
	return nil
}
func (m *draftMockRatingSvc) DeleteGroupRatings(ctx context.Context, groupID int64) error {
	return nil
}
func (m *draftMockRatingSvc) RecalculateAllRatings(ctx context.Context) (RecalcResult, error) {
	return RecalcResult{}, nil
}

func TestFinishGroup_DNSMatchExemptFromDoneRequirement(t *testing.T) {
	// Setup: group with 3 players. p3 is DNS. p1 vs p2 is DONE. p1 vs p3 and p2 vs p3 are DRAFT.
	// FinishGroup should succeed because DNS player matches are exempt.
	gp1, gp2, gp3 := int64(1), int64(2), int64(3)
	players := []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 10, UserID: 1, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: gp2, GroupID: 10, UserID: 2, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: gp3, GroupID: 10, UserID: 3, PlayerStatus: model.PlayerStatusDNS},
	}

	s3, s1 := int16(3), int16(1)
	matches := []model.Match{
		{MatchID: 1, GroupID: 10, GroupPlayer1ID: &gp1, GroupPlayer2ID: &gp2, Score1: &s3, Score2: &s1, Status: model.MatchDone},
		{MatchID: 2, GroupID: 10, GroupPlayer1ID: &gp1, GroupPlayer2ID: &gp3, Status: model.MatchDraft}, // exempt
		{MatchID: 3, GroupID: 10, GroupPlayer1ID: &gp2, GroupPlayer2ID: &gp3, Status: model.MatchDraft}, // exempt
	}

	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{
			10: {GroupID: 10, EventID: 1, Status: model.GroupInProgress},
		},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	gr.groups = map[int64][]model.Group{1: {{GroupID: 10, EventID: 1, Status: model.GroupInProgress}}}

	mr := &draftMockMatchRepo{matches: map[int64][]model.Match{10: matches}}

	er := &draftMockEventRepo{
		events: map[int64]*model.LeagueEvent{
			1: {EventID: 1, Status: model.EventInProgress, LeagueID: 100},
		},
	}

	leagueRepo := &draftMockLeagueRepo{leagues: map[int64]*model.League{
		100: {LeagueID: 100, Config: model.LeagueConfig{NumberOfAdvances: 1, NumberOfRecedes: 1, GamesToWin: 3}},
	}}
	matchSvc := &draftMockMatchSvc{}
	groupSvc := &draftMockGroupSvc{}
	ratingSvc := &draftMockRatingSvc{}

	svc := &draftService{
		groupRepo:  gr,
		matchRepo:  mr,
		eventRepo:  er,
		leagueRepo: leagueRepo,
		matchSvc:   matchSvc,
		groupSvc:   groupSvc,
		ratingSvc:  ratingSvc,
	}

	err := svc.FinishGroup(context.Background(), 10)
	if err != nil {
		t.Fatalf("expected FinishGroup to succeed with DNS exempt matches, got: %v", err)
	}
	if !matchSvc.recalcCalled {
		t.Error("expected RecalcGroupPoints to be called")
	}
}

func TestFinishGroup_NonDNSUnfinishedMatchBlocks(t *testing.T) {
	// Both players are active. A DRAFT match should block FinishGroup.
	gp1, gp2 := int64(1), int64(2)
	players := []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 10, UserID: 1, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: gp2, GroupID: 10, UserID: 2, PlayerStatus: model.PlayerStatusActive},
	}

	matches := []model.Match{
		{MatchID: 1, GroupID: 10, GroupPlayer1ID: &gp1, GroupPlayer2ID: &gp2, Status: model.MatchDraft},
	}

	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{
			10: {GroupID: 10, EventID: 1, Status: model.GroupInProgress},
		},
		players: map[int64][]model.GroupPlayer{10: players},
	}

	mr := &draftMockMatchRepo{matches: map[int64][]model.Match{10: matches}}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{1: inProgressEvent(1)}}

	svc := &draftService{
		groupRepo: gr,
		matchRepo: mr,
		eventRepo: er,
	}

	err := svc.FinishGroup(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error: unfinished non-DNS match should block FinishGroup")
	}
	if !strings.Contains(err.Error(), "no score yet") {
		t.Errorf("unexpected error message: %v", err)
	}
}

