package service

import (
	"context"
	"errors"
	"testing"

	"league-api/internal/model"
)

// errGroupRepo wraps mockGroupRepo but injects errors at specific calls.
type errGroupRepo struct {
	mockGroupRepo
	getByIDErr  error
	getPlayersErr error
}

func (m *errGroupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.mockGroupRepo.GetByID(ctx, id)
}

func (m *errGroupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	if m.getPlayersErr != nil {
		return nil, m.getPlayersErr
	}
	return m.mockGroupRepo.GetPlayers(ctx, groupID)
}

// errMatchRepo returns an error from ListByGroup.
type errMatchRepo struct {
	mockMatchRepo
	listErr error
}

func (m *errMatchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.mockMatchRepo.ListByGroup(ctx, groupID)
}

// --- GetGroupDetail error paths ---

func TestGetGroupDetail_GetByIDError(t *testing.T) {
	gr := &errGroupRepo{
		mockGroupRepo: mockGroupRepo{players: map[int64][]model.GroupPlayer{}},
		getByIDErr:    errors.New("db error"),
	}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{}}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	_, _, _, err := svc.GetGroupDetail(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from GetByID")
	}
}

func TestGetGroupDetail_GetPlayersError(t *testing.T) {
	gr := &errGroupRepo{
		mockGroupRepo:  mockGroupRepo{players: map[int64][]model.GroupPlayer{}},
		getPlayersErr: errors.New("db error"),
	}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{}}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	_, _, _, err := svc.GetGroupDetail(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from GetPlayers")
	}
}

func TestGetGroupDetail_ListByGroupError(t *testing.T) {
	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: {}}}
	mr := &errMatchRepo{
		mockMatchRepo: mockMatchRepo{matches: map[int64][]model.Match{}},
		listErr:       errors.New("db error"),
	}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	_, _, _, err := svc.GetGroupDetail(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from ListByGroup")
	}
}

// --- headToHeadWinner withdraw paths ---

func TestHeadToHeadWinner_Withdraw1_AsP1(t *testing.T) {
	// p1 is GroupPlayer1 and withdraws → p2 wins.
	gp1, gp2 := int64(1), int64(2)
	m := model.Match{
		MatchID:        1,
		GroupPlayer1ID: &gp1,
		GroupPlayer2ID: &gp2,
		Withdraw1:      true,
		Status:         model.MatchDone,
	}
	winner := headToHeadWinner(1, 2, []model.Match{m})
	if winner != 2 {
		t.Errorf("expected p2 to win when p1 withdraws (p1 is GroupPlayer1), got %d", winner)
	}
}

func TestHeadToHeadWinner_Withdraw2_AsP1(t *testing.T) {
	// p1 is GroupPlayer1, p2 withdraws → p1 wins.
	gp1, gp2 := int64(1), int64(2)
	m := model.Match{
		MatchID:        1,
		GroupPlayer1ID: &gp1,
		GroupPlayer2ID: &gp2,
		Withdraw2:      true,
		Status:         model.MatchDone,
	}
	winner := headToHeadWinner(1, 2, []model.Match{m})
	if winner != 1 {
		t.Errorf("expected p1 to win when p2 withdraws (p1 is GroupPlayer1), got %d", winner)
	}
}

func TestHeadToHeadWinner_Withdraw1_AsP2(t *testing.T) {
	// p2 is GroupPlayer1 (p1 is GroupPlayer2); GroupPlayer1 (p2) withdraws → p1 wins.
	gp1, gp2 := int64(2), int64(1) // p2 is player1, p1 is player2 in match
	m := model.Match{
		MatchID:        1,
		GroupPlayer1ID: &gp1,
		GroupPlayer2ID: &gp2,
		Withdraw1:      true,
		Status:         model.MatchDone,
	}
	winner := headToHeadWinner(1, 2, []model.Match{m})
	// p2IsPlayer1 → Withdraw1 → return p1ID (the argument p1ID=1)
	if winner != 1 {
		t.Errorf("expected p1 to win when p2 (GroupPlayer1) withdraws, got %d", winner)
	}
}

func TestHeadToHeadWinner_Withdraw2_AsP2(t *testing.T) {
	// p2 is GroupPlayer1, p1 is GroupPlayer2; GroupPlayer2 (p1) withdraws → p2 wins.
	gp1, gp2 := int64(2), int64(1) // p2 is player1, p1 is player2 in match
	m := model.Match{
		MatchID:        1,
		GroupPlayer1ID: &gp1,
		GroupPlayer2ID: &gp2,
		Withdraw2:      true,
		Status:         model.MatchDone,
	}
	winner := headToHeadWinner(1, 2, []model.Match{m})
	if winner != 2 {
		t.Errorf("expected p2 to win when p1 (GroupPlayer2) withdraws, got %d", winner)
	}
}

func TestHeadToHeadWinner_P2WinsFromP2Position(t *testing.T) {
	// p1 is GroupPlayer2, p2 is GroupPlayer1; p2 wins (higher score).
	gp1, gp2 := int64(2), int64(1) // gp1=p2, gp2=p1 as args
	s1, s2 := int16(3), int16(1)
	m := model.Match{
		MatchID:        1,
		GroupPlayer1ID: &gp1,
		GroupPlayer2ID: &gp2,
		Score1:         &s1,
		Score2:         &s2,
		Status:         model.MatchDone,
	}
	// headToHeadWinner(p1ID=1, p2ID=2, matches)
	// p1IsPlayer2: GroupPlayer1=2=p2ID, GroupPlayer2=1=p1ID → p1IsPlayer2 = true
	// Score1=3 > Score2=1 → return p2ID = 2
	winner := headToHeadWinner(1, 2, []model.Match{m})
	if winner != 2 {
		t.Errorf("expected p2 to win (3:1 as GroupPlayer1), got %d", winner)
	}
}

func TestHeadToHeadWinner_P1WinsFromP2Position(t *testing.T) {
	// p1 is GroupPlayer2, p2 is GroupPlayer1; p1 wins.
	gp1, gp2 := int64(2), int64(1) // gp1=p2, gp2=p1 in match
	s1, s2 := int16(1), int16(3)   // GroupPlayer1(p2) scores 1, GroupPlayer2(p1) scores 3
	m := model.Match{
		MatchID:        1,
		GroupPlayer1ID: &gp1,
		GroupPlayer2ID: &gp2,
		Score1:         &s1,
		Score2:         &s2,
		Status:         model.MatchDone,
	}
	// p1IsPlayer2: true; Score1(1) < Score2(3) → return p1ID = 1
	winner := headToHeadWinner(1, 2, []model.Match{m})
	if winner != 1 {
		t.Errorf("expected p1 to win (1:3 as GroupPlayer2), got %d", winner)
	}
}

// --- SeedPlayer - player already in event ---

// seedGroupRepoWithExisting wraps mockGroupRepo and returns existing players for ListPlayerGroupsInEvent.
type seedGroupRepoWithExisting struct {
	mockGroupRepo
	existingPlayers []model.GroupPlayer
}

func (m *seedGroupRepoWithExisting) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return m.existingPlayers, nil
}

func TestSeedPlayer_AlreadyInEventError(t *testing.T) {
	gr := &seedGroupRepoWithExisting{
		mockGroupRepo: mockGroupRepo{players: map[int64][]model.GroupPlayer{1: {}}},
		existingPlayers: []model.GroupPlayer{
			{GroupPlayerID: 99, GroupID: 2, UserID: 42}, // player 42 already in event
		},
	}
	mr := &mockMatchRepo{}
	er := &mockDraftEventRepo{} // returns DRAFT events

	svc := NewGroupService(nil,gr, mr, er, nil)
	err := svc.SeedPlayer(context.Background(), 1, 42)
	if err == nil {
		t.Fatal("expected error for player already in event")
	}
}

// --- finaliseGroup error paths ---

func TestFinaliseGroup_RatingCalcError(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
		players:   map[int64][]model.GroupPlayer{},
	}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}

	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
		ratingSvc: &nopRatingService{calcErr: errors.New("calc error")},
	}

	err := svc.finaliseGroup(context.Background(), grp)
	if err == nil {
		t.Fatal("expected error from CalculateGroupRatings")
	}
}

// --- FinishGroup — nil hub path (broadcasts to nil hub) ---

func TestFinishGroup_ManualPlacementsRequired_WithNilHub(t *testing.T) {
	grp := &model.Group{GroupID: 1, EventID: 10, Status: model.GroupInProgress}
	gr := &draftMockGroupRepo{
		groupByID: map[int64]*model.Group{1: grp},
		groups:    map[int64][]model.Group{},
		players:   map[int64][]model.GroupPlayer{},
	}
	mr := &draftMockMatchRepo{matches: map[int64][]model.Match{1: {}}}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}

	groupSvc := &nopGroupService{needsManual: []int64{1, 2, 3}}
	svc := &draftService{
		groupRepo: gr,
		eventRepo: er,
		matchRepo: mr,
		matchSvc:  &nopMatchService{},
		groupSvc:  groupSvc,
		hub:       nil, // nil hub — should not panic
	}

	err := svc.FinishGroup(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil error for manual placement required with nil hub, got: %v", err)
	}
}
