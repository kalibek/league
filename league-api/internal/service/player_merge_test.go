package service

import (
	"context"
	"errors"
	"testing"

	"league-api/internal/model"
	"league-api/internal/repository"
)

// --- merge test mocks ---

type mergeUserRepo struct {
	users         map[int64]*model.User
	softDeleted   map[int64]int64 // sourceID → targetID
	historyMoved  []int64         // sourceIDs whose history was moved
}

func newMergeUserRepo(users ...*model.User) *mergeUserRepo {
	m := &mergeUserRepo{
		users:        make(map[int64]*model.User),
		softDeleted:  make(map[int64]int64),
		historyMoved: nil,
	}
	for _, u := range users {
		m.users[u.UserID] = u
	}
	return m
}

func (m *mergeUserRepo) FindAllActive(_ context.Context) ([]model.User, error) {
	out := make([]model.User, 0, len(m.users))
	for _, u := range m.users {
		if u.MergedIntoUserID == nil {
			out = append(out, *u)
		}
	}
	return out, nil
}
func (m *mergeUserRepo) SoftDeleteMerged(_ context.Context, sourceID, targetID int64) error {
	m.softDeleted[sourceID] = targetID
	id := targetID
	m.users[sourceID].MergedIntoUserID = &id
	return nil
}
func (m *mergeUserRepo) UpdateRatingHistory(_ context.Context, sourceID, _ int64) error {
	m.historyMoved = append(m.historyMoved, sourceID)
	return nil
}
func (m *mergeUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}
func (m *mergeUserRepo) GetByEmail(_ context.Context, _ string) (*model.User, error) { return nil, errors.New("not found") }
func (m *mergeUserRepo) Create(_ context.Context, _ *model.User) (int64, error)       { return 0, nil }
func (m *mergeUserRepo) List(_ context.Context, _, _ int, _ string) ([]model.User, error) {
	return nil, nil
}
func (m *mergeUserRepo) Search(_ context.Context, _ string, _, _ int, _ string) ([]model.User, error) {
	return nil, nil
}
func (m *mergeUserRepo) UpdateRating(_ context.Context, _ int64, _, _, _ float64) error { return nil }
func (m *mergeUserRepo) ResetAllRatings(_ context.Context) error                         { return nil }
func (m *mergeUserRepo) SetPasswordHash(_ context.Context, _ int64, _ string) error      { return nil }
func (m *mergeUserRepo) UpdateName(_ context.Context, _ int64, _, _ string) error        { return nil }

type mergeGroupRepo struct {
	conflicts        []int64 // returned by FindConflictingGroupIDs
	dnsSet           []int64 // groupIDs where DNS was set
	gpUserIDUpdated  bool
	profilesDeleted  []int64
}

func (m *mergeGroupRepo) FindConflictingGroupIDs(_ context.Context, _, _ int64) ([]int64, error) {
	return m.conflicts, nil
}
func (m *mergeGroupRepo) SetPlayerStatusByUser(_ context.Context, groupID, _ int64, _ model.PlayerStatus) error {
	m.dnsSet = append(m.dnsSet, groupID)
	return nil
}
func (m *mergeGroupRepo) UpdateGroupPlayerUserID(_ context.Context, _, _ int64, _ []int64) error {
	m.gpUserIDUpdated = true
	return nil
}
func (m *mergeGroupRepo) DeletePlayerProfileByUser(_ context.Context, userID int64) error {
	m.profilesDeleted = append(m.profilesDeleted, userID)
	return nil
}

// Remaining GroupRepository stubs
func (m *mergeGroupRepo) GetByID(_ context.Context, _ int64) (*model.Group, error) { return nil, nil }
func (m *mergeGroupRepo) ListByEvent(_ context.Context, _ int64) ([]model.Group, error) {
	return nil, nil
}
func (m *mergeGroupRepo) Create(_ context.Context, _ *model.Group) (int64, error) { return 0, nil }
func (m *mergeGroupRepo) UpdateStatus(_ context.Context, _ int64, _ model.GroupStatus) error {
	return nil
}
func (m *mergeGroupRepo) GetPlayers(_ context.Context, _ int64) ([]model.GroupPlayer, error) {
	return nil, nil
}
func (m *mergeGroupRepo) GetPlayersByMovement(_ context.Context, _ int64, _ int) ([]model.GroupPlayer, error) {
	return nil, nil
}
func (m *mergeGroupRepo) AddPlayer(_ context.Context, _ *model.GroupPlayer) (int64, error) {
	return 0, nil
}
func (m *mergeGroupRepo) UpdatePlayer(_ context.Context, _ *model.GroupPlayer) error { return nil }
func (m *mergeGroupRepo) SetPlayerStatus(_ context.Context, _ int64, _ model.PlayerStatus) error {
	return nil
}
func (m *mergeGroupRepo) RemovePlayer(_ context.Context, _ int64) error           { return nil }
func (m *mergeGroupRepo) ResetGroupPlayers(_ context.Context, _ int64) error      { return nil }
func (m *mergeGroupRepo) ListPlayerGroupsInEvent(_ context.Context, _, _ int64) ([]model.GroupPlayer, error) {
	return nil, nil
}
func (m *mergeGroupRepo) ListUsersByIdsByRatingDesc(_ context.Context, _ []int64) ([]model.User, error) {
	return nil, nil
}
func (m *mergeGroupRepo) Delete(_ context.Context, _ int64) error { return nil }

type mergeRatingRepo struct {
	earliestEventID int64
	hasEvent        bool
}

func (m *mergeRatingRepo) GetEarliestEventIDForUser(_ context.Context, _ int64) (int64, bool, error) {
	return m.earliestEventID, m.hasEvent, nil
}
func (m *mergeRatingRepo) InsertHistory(_ context.Context, _ *model.RatingHistory) error { return nil }
func (m *mergeRatingRepo) GetByUser(_ context.Context, _ int64) ([]model.RatingHistory, error) {
	return nil, nil
}
func (m *mergeRatingRepo) GetByUserInEvent(_ context.Context, _, _ int64) ([]model.RatingHistory, error) {
	return nil, nil
}
func (m *mergeRatingRepo) DeleteByGroup(_ context.Context, _ int64) error             { return nil }
func (m *mergeRatingRepo) DeleteAll(_ context.Context) error                           { return nil }
func (m *mergeRatingRepo) GetEventDeltaForUser(_ context.Context, _, _ int64) (float64, error) {
	return 0, nil
}
func (m *mergeRatingRepo) DeleteFromEvent(_ context.Context, _ int64) error { return nil }
func (m *mergeRatingRepo) GetLastRatingsBeforeEvent(_ context.Context, _ int64) ([]model.UserRatingSnapshot, error) {
	return nil, nil
}

type mergeRatingSvc struct {
	calledWithEvent int64
}

func (m *mergeRatingSvc) RecalculateFromEvent(_ context.Context, fromEventID int64) (RecalcResult, error) {
	m.calledWithEvent = fromEventID
	return RecalcResult{EventsProcessed: 1}, nil
}
func (m *mergeRatingSvc) CalculateGroupRatings(_ context.Context, _ int64) error { return nil }
func (m *mergeRatingSvc) RecalculateGroupRatings(_ context.Context, _ int64) error { return nil }
func (m *mergeRatingSvc) DeleteGroupRatings(_ context.Context, _ int64) error      { return nil }
func (m *mergeRatingSvc) RecalculateAllRatings(_ context.Context) (RecalcResult, error) {
	return RecalcResult{}, nil
}

func newMergeSvc(ur repository.UserRepository, gr repository.GroupRepository, rr repository.RatingRepository, rs RatingService) PlayerService {
	return &playerService{
		userRepo:   ur,
		ratingRepo: rr,
		groupRepo:  gr,
		ratingSvc:  rs,
	}
}

// --- FindDuplicates tests ---

func TestFindDuplicates_DetectsKazakhVariants(t *testing.T) {
	ur := newMergeUserRepo(
		&model.User{UserID: 1, FirstName: "Куаныш", LastName: "Тынышбай"},
		&model.User{UserID: 2, FirstName: "Қуаныш", LastName: "Тынышбай"},
		&model.User{UserID: 3, FirstName: "Алибек", LastName: "Сейткали"},
	)
	svc := newMergeSvc(ur, &mergeGroupRepo{}, &mergeRatingRepo{}, &mergeRatingSvc{})

	groups, err := svc.FindDuplicates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}
	if len(groups[0].Users) != 2 {
		t.Errorf("expected 2 users in group, got %d", len(groups[0].Users))
	}
}

func TestFindDuplicates_NoDuplicates(t *testing.T) {
	ur := newMergeUserRepo(
		&model.User{UserID: 1, FirstName: "Алибек", LastName: "Сейткали"},
		&model.User{UserID: 2, FirstName: "Марат", LastName: "Жумабеков"},
	)
	svc := newMergeSvc(ur, &mergeGroupRepo{}, &mergeRatingRepo{}, &mergeRatingSvc{})

	groups, err := svc.FindDuplicates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

// --- MergeUsers tests ---

func TestMergeUsers_NoConflict(t *testing.T) {
	ur := newMergeUserRepo(
		&model.User{UserID: 10, FirstName: "Alice", LastName: "Smith"},
		&model.User{UserID: 20, FirstName: "Alice", LastName: "Smith"},
	)
	gr := &mergeGroupRepo{conflicts: nil}
	rr := &mergeRatingRepo{earliestEventID: 5, hasEvent: true}
	rs := &mergeRatingSvc{}

	svc := newMergeSvc(ur, gr, rr, rs)
	result, err := svc.MergeUsers(context.Background(), 10, []int64{20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TargetID != 10 {
		t.Errorf("expected targetID 10, got %d", result.TargetID)
	}
	if len(result.MergedSourceIDs) != 1 || result.MergedSourceIDs[0] != 20 {
		t.Errorf("expected mergedSourceIDs=[20], got %v", result.MergedSourceIDs)
	}
	// Soft delete happened.
	if targetID, ok := ur.softDeleted[20]; !ok || targetID != 10 {
		t.Errorf("expected soft delete of 20→10, got %v", ur.softDeleted)
	}
	// Rating history moved.
	if len(ur.historyMoved) != 1 || ur.historyMoved[0] != 20 {
		t.Errorf("expected historyMoved=[20], got %v", ur.historyMoved)
	}
	// No conflict groups → DNS never set.
	if len(gr.dnsSet) != 0 {
		t.Errorf("expected no DNS sets, got %v", gr.dnsSet)
	}
	// UpdateGroupPlayerUserID called.
	if !gr.gpUserIDUpdated {
		t.Error("expected UpdateGroupPlayerUserID to be called")
	}
	// Recalc triggered.
	if rs.calledWithEvent != 5 {
		t.Errorf("expected recalc from event 5, got %d", rs.calledWithEvent)
	}
	if result.RecalcFromEvent == nil || *result.RecalcFromEvent != 5 {
		t.Errorf("expected RecalcFromEvent=5, got %v", result.RecalcFromEvent)
	}
}

func TestMergeUsers_WithConflict(t *testing.T) {
	ur := newMergeUserRepo(
		&model.User{UserID: 10, FirstName: "Alice", LastName: "Smith"},
		&model.User{UserID: 20, FirstName: "Alice", LastName: "Smith"},
	)
	gr := &mergeGroupRepo{conflicts: []int64{99}} // group 99 has both
	rr := &mergeRatingRepo{hasEvent: false}
	rs := &mergeRatingSvc{}

	svc := newMergeSvc(ur, gr, rr, rs)
	result, err := svc.MergeUsers(context.Background(), 10, []int64{20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// DNS was set for conflict group.
	if len(gr.dnsSet) != 1 || gr.dnsSet[0] != 99 {
		t.Errorf("expected DNS set for group 99, got %v", gr.dnsSet)
	}
	if len(result.ConflictGroups) != 1 || result.ConflictGroups[0] != 99 {
		t.Errorf("expected ConflictGroups=[99], got %v", result.ConflictGroups)
	}
	// No rating history → no recalc.
	if rs.calledWithEvent != 0 {
		t.Errorf("expected no recalc, got calledWithEvent=%d", rs.calledWithEvent)
	}
}

func TestMergeUsers_AlreadyMerged(t *testing.T) {
	mergedID := int64(10)
	ur := newMergeUserRepo(
		&model.User{UserID: 10, FirstName: "Alice", LastName: "Smith"},
		&model.User{UserID: 20, FirstName: "Alice", LastName: "Smith", MergedIntoUserID: &mergedID},
	)
	svc := newMergeSvc(ur, &mergeGroupRepo{}, &mergeRatingRepo{}, &mergeRatingSvc{})

	_, err := svc.MergeUsers(context.Background(), 10, []int64{20})
	if err == nil {
		t.Fatal("expected error for already-merged source")
	}
}

func TestMergeUsers_SourceEqualsTarget(t *testing.T) {
	ur := newMergeUserRepo(
		&model.User{UserID: 10, FirstName: "Alice", LastName: "Smith"},
	)
	svc := newMergeSvc(ur, &mergeGroupRepo{}, &mergeRatingRepo{}, &mergeRatingSvc{})

	_, err := svc.MergeUsers(context.Background(), 10, []int64{10})
	if err == nil {
		t.Fatal("expected error when sourceID equals targetID")
	}
}
