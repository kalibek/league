package service

import (
	"context"
	"errors"
	"strings"
	"testing"

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
	return nil, nil
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
