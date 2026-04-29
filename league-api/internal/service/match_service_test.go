package service

import (
	"context"
	"errors"
	"testing"

	"league-api/internal/model"
	"league-api/internal/ws"
)

// newTestHub creates a Hub with its Run goroutine started in background.
func newTestHub() *ws.Hub {
	h := ws.NewHub()
	go h.Run()
	return h
}

// --- match service mocks ---

type matchSvcMockMatchRepo struct {
	matches      map[int64]*model.Match
	groupMatches map[int64][]model.Match
	scoreErr     error
	statusErr    error
	updateCalls  int
	statusCalls  int
}

func (m *matchSvcMockMatchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) {
	if match, ok := m.matches[id]; ok {
		return match, nil
	}
	return nil, errors.New("match not found")
}

func (m *matchSvcMockMatchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	return m.groupMatches[groupID], nil
}

func (m *matchSvcMockMatchRepo) Create(ctx context.Context, match *model.Match) (int64, error) {
	return 1, nil
}

func (m *matchSvcMockMatchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error {
	m.updateCalls++
	if m.scoreErr != nil {
		return m.scoreErr
	}
	if match, ok := m.matches[id]; ok {
		match.Score1 = &score1
		match.Score2 = &score2
		match.Withdraw1 = withdraw1
		match.Withdraw2 = withdraw2
	}
	return nil
}

func (m *matchSvcMockMatchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	m.statusCalls++
	if m.statusErr != nil {
		return m.statusErr
	}
	if match, ok := m.matches[id]; ok {
		match.Status = status
	}
	return nil
}

func (m *matchSvcMockMatchRepo) BulkCreate(ctx context.Context, matches []model.Match) error { return nil }

func (m *matchSvcMockMatchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error { return nil }

func (m *matchSvcMockMatchRepo) SetWithdraw(ctx context.Context, matchID int64, position int) error {
	return nil
}

func (m *matchSvcMockMatchRepo) SetTableNumber(ctx context.Context, matchID int64, tableNumber int) error {
	if match, ok := m.matches[matchID]; ok {
		match.TableNumber = &tableNumber
		match.Status = model.MatchInProgress
	}
	return nil
}

func (m *matchSvcMockMatchRepo) ListInProgressByEvent(ctx context.Context, eventID int64) ([]model.Match, error) {
	var result []model.Match
	for _, match := range m.matches {
		if match.Status == model.MatchInProgress {
			result = append(result, *match)
		}
	}
	return result, nil
}

type matchSvcMockGroupRepo struct {
	groups  map[int64]*model.Group
	players map[int64][]model.GroupPlayer
	updated []model.GroupPlayer
}

func (m *matchSvcMockGroupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	if g, ok := m.groups[id]; ok {
		return g, nil
	}
	return nil, errors.New("group not found")
}

func (m *matchSvcMockGroupRepo) ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error) {
	return nil, nil
}

func (m *matchSvcMockGroupRepo) Create(ctx context.Context, g *model.Group) (int64, error) { return 1, nil }

func (m *matchSvcMockGroupRepo) UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error {
	return nil
}

func (m *matchSvcMockGroupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	return m.players[groupID], nil
}

func (m *matchSvcMockGroupRepo) GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error) {
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

func (m *matchSvcMockGroupRepo) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	return 1, nil
}

func (m *matchSvcMockGroupRepo) UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error {
	m.updated = append(m.updated, *gp)
	players := m.players[gp.GroupID]
	for i, p := range players {
		if p.GroupPlayerID == gp.GroupPlayerID {
			m.players[gp.GroupID][i] = *gp
		}
	}
	return nil
}

func (m *matchSvcMockGroupRepo) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }

func (m *matchSvcMockGroupRepo) ResetGroupPlayers(ctx context.Context, groupID int64) error { return nil }

func (m *matchSvcMockGroupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *matchSvcMockGroupRepo) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	users := make([]model.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, model.User{UserID: id})
	}
	return users, nil
}

// buildMatch creates a DRAFT match between two group players in a group.
func buildMatch(matchID, groupID, gp1, gp2 int64) *model.Match {
	return &model.Match{
		MatchID:        matchID,
		GroupID:        groupID,
		GroupPlayer1ID: &gp1,
		GroupPlayer2ID: &gp2,
		Status:         model.MatchDraft,
	}
}

func buildGroup(groupID, eventID int64) *model.Group {
	return &model.Group{GroupID: groupID, EventID: eventID, Status: model.GroupInProgress}
}

// buildGroupPlayers creates two players in a group with the given groupPlayerIDs.
func buildGroupPlayers(groupID int64, ids []int64) []model.GroupPlayer {
	result := make([]model.GroupPlayer, len(ids))
	for i, id := range ids {
		result[i] = model.GroupPlayer{
			GroupPlayerID: id,
			GroupID:       groupID,
			UserID:        id,
		}
	}
	return result
}

// --- UpdateScore tests ---

func TestUpdateScore_Valid_p1wins(t *testing.T) {
	match := buildMatch(1, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches:      map[int64]*model.Match{1: match},
		groupMatches: map[int64][]model.Match{10: {*match}},
	}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: buildGroupPlayers(10, []int64{1, 2})},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr, hub: newTestHub()}

	err := svc.UpdateScore(context.Background(), 1, 3, 1, 3, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.updateCalls != 1 {
		t.Errorf("expected UpdateScore called once, got %d", mr.updateCalls)
	}
	if mr.statusCalls != 1 {
		t.Errorf("expected UpdateStatus called once, got %d", mr.statusCalls)
	}
}

func TestUpdateScore_Valid_p2wins(t *testing.T) {
	match := buildMatch(2, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches:      map[int64]*model.Match{2: match},
		groupMatches: map[int64][]model.Match{10: {*match}},
	}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: buildGroupPlayers(10, []int64{1, 2})},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr, hub: newTestHub()}

	err := svc.UpdateScore(context.Background(), 2, 1, 3, 3, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateScore_InvalidScore_BothEqual(t *testing.T) {
	match := buildMatch(1, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches: map[int64]*model.Match{1: match},
	}
	gr := &matchSvcMockGroupRepo{}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.UpdateScore(context.Background(), 1, 3, 3, 3, false, false)
	if err == nil {
		t.Fatal("expected error for equal scores")
	}
}

func TestUpdateScore_InvalidScore_BothLessThanGTW(t *testing.T) {
	match := buildMatch(1, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches: map[int64]*model.Match{1: match},
	}
	gr := &matchSvcMockGroupRepo{}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.UpdateScore(context.Background(), 1, 2, 1, 3, false, false)
	if err == nil {
		t.Fatal("expected error when neither score equals gamesToWin")
	}
}

func TestUpdateScore_InvalidScore_BothExceedGTW(t *testing.T) {
	match := buildMatch(1, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches: map[int64]*model.Match{1: match},
	}
	gr := &matchSvcMockGroupRepo{}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	// score1 > gamesToWin — invalid
	err := svc.UpdateScore(context.Background(), 1, 4, 3, 3, false, false)
	if err == nil {
		t.Fatal("expected error when score1 > gamesToWin")
	}
}

func TestUpdateScore_MatchNotFound(t *testing.T) {
	mr := &matchSvcMockMatchRepo{
		matches: map[int64]*model.Match{},
	}
	gr := &matchSvcMockGroupRepo{}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.UpdateScore(context.Background(), 99, 3, 1, 3, false, false)
	if err == nil {
		t.Fatal("expected error for missing match")
	}
}

func TestUpdateScore_UpdateScoreRepoError(t *testing.T) {
	match := buildMatch(1, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches:  map[int64]*model.Match{1: match},
		scoreErr: errors.New("db error"),
	}
	gr := &matchSvcMockGroupRepo{}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.UpdateScore(context.Background(), 1, 3, 1, 3, false, false)
	if err == nil {
		t.Fatal("expected error from UpdateScore repo")
	}
}

func TestUpdateScore_UpdateStatusRepoError(t *testing.T) {
	match := buildMatch(1, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches:   map[int64]*model.Match{1: match},
		statusErr: errors.New("status db error"),
	}
	gr := &matchSvcMockGroupRepo{}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.UpdateScore(context.Background(), 1, 3, 1, 3, false, false)
	if err == nil {
		t.Fatal("expected error from UpdateStatus repo")
	}
}

// --- recalcGroupPoints tests ---

func TestRecalcGroupPoints_AllMatchesDone(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)
	matches := []model.Match{
		{
			MatchID:        1,
			GroupID:        10,
			GroupPlayer1ID: &gp1,
			GroupPlayer2ID: &gp2,
			Score1:         &s1,
			Score2:         &s2,
			Status:         model.MatchDone,
		},
	}
	players := buildGroupPlayers(10, []int64{1, 2})

	mr := &matchSvcMockMatchRepo{groupMatches: map[int64][]model.Match{10: matches}}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.recalcGroupPoints(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// player 1 (gp1=1) wins: should have 2 pts; player 2 (gp2=2) loses: 1 pt
	pts1 := gr.players[10][0].Points
	pts2 := gr.players[10][1].Points
	if pts1 != 2 {
		t.Errorf("expected player 1 to have 2 points, got %d", pts1)
	}
	if pts2 != 1 {
		t.Errorf("expected player 2 to have 1 point, got %d", pts2)
	}
}

func TestRecalcGroupPoints_WithdrawPlayer1(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	// Score pointers must be non-nil because the points pass checks nil scores first.
	s0 := int16(0)
	matches := []model.Match{
		{
			MatchID:        1,
			GroupID:        10,
			GroupPlayer1ID: &gp1,
			GroupPlayer2ID: &gp2,
			Score1:         &s0,
			Score2:         &s0,
			Withdraw1:      true,
			Status:         model.MatchDone,
		},
	}
	players := buildGroupPlayers(10, []int64{1, 2})

	mr := &matchSvcMockMatchRepo{groupMatches: map[int64][]model.Match{10: matches}}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.recalcGroupPoints(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// player 2 gets 2 points (withdrawal by p1); player 1 gets 0
	pts2 := gr.players[10][1].Points
	if pts2 != 2 {
		t.Errorf("expected player 2 to have 2 points on p1 withdrawal, got %d", pts2)
	}
}

func TestRecalcGroupPoints_WithdrawPlayer2(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s0 := int16(0)
	matches := []model.Match{
		{
			MatchID:        1,
			GroupID:        10,
			GroupPlayer1ID: &gp1,
			GroupPlayer2ID: &gp2,
			Score1:         &s0,
			Score2:         &s0,
			Withdraw2:      true,
			Status:         model.MatchDone,
		},
	}
	players := buildGroupPlayers(10, []int64{1, 2})

	mr := &matchSvcMockMatchRepo{groupMatches: map[int64][]model.Match{10: matches}}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.recalcGroupPoints(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pts1 := gr.players[10][0].Points
	if pts1 != 2 {
		t.Errorf("expected player 1 to have 2 points on p2 withdrawal, got %d", pts1)
	}
}

func TestRecalcGroupPoints_SkipsNonDoneMatches(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	matches := []model.Match{
		{
			MatchID:        1,
			GroupID:        10,
			GroupPlayer1ID: &gp1,
			GroupPlayer2ID: &gp2,
			Status:         model.MatchDraft, // not done
		},
	}
	players := buildGroupPlayers(10, []int64{1, 2})

	mr := &matchSvcMockMatchRepo{groupMatches: map[int64][]model.Match{10: matches}}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.recalcGroupPoints(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No completed matches so everyone has 0 points
	for _, p := range gr.players[10] {
		if p.Points != 0 {
			t.Errorf("expected 0 points for player %d, got %d", p.GroupPlayerID, p.Points)
		}
	}
}

func TestRecalcGroupPoints_TiebreakPoints(t *testing.T) {
	// Three players, all tied on points (1 win 1 loss = 3 pts each).
	// p1 beats p2 3:1, p2 beats p3 3:0, p3 beats p1 3:2.
	gp1, gp2, gp3 := int64(1), int64(2), int64(3)
	s31, s13 := int16(3), int16(1)
	s30, s00 := int16(3), int16(0)
	s32, s23 := int16(3), int16(2)

	matches := []model.Match{
		{MatchID: 1, GroupID: 10, GroupPlayer1ID: &gp1, GroupPlayer2ID: &gp2, Score1: &s31, Score2: &s13, Status: model.MatchDone},
		{MatchID: 2, GroupID: 10, GroupPlayer1ID: &gp2, GroupPlayer2ID: &gp3, Score1: &s30, Score2: &s00, Status: model.MatchDone},
		{MatchID: 3, GroupID: 10, GroupPlayer1ID: &gp3, GroupPlayer2ID: &gp1, Score1: &s32, Score2: &s23, Status: model.MatchDone},
	}
	players := buildGroupPlayers(10, []int64{1, 2, 3})

	mr := &matchSvcMockMatchRepo{groupMatches: map[int64][]model.Match{10: matches}}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.recalcGroupPoints(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All 3 players should have 3 points (1 win + 1 loss).
	for _, p := range gr.players[10] {
		if p.Points != 3 {
			t.Errorf("player %d expected 3 points, got %d", p.GroupPlayerID, p.Points)
		}
	}
}

func TestRecalcGroupPoints_NonCalculatedPlayerExcludedFromTiebreak(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s3, s1 := int16(3), int16(1)
	matches := []model.Match{
		{MatchID: 1, GroupID: 10, GroupPlayer1ID: &gp1, GroupPlayer2ID: &gp2, Score1: &s3, Score2: &s1, Status: model.MatchDone},
	}
	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 10, UserID: 1},
		{GroupPlayerID: 2, GroupID: 10, UserID: 2, IsNonCalculated: true},
	}

	mr := &matchSvcMockMatchRepo{groupMatches: map[int64][]model.Match{10: matches}}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.recalcGroupPoints(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- SetTableNumber tests ---

func TestSetTableNumber_Success(t *testing.T) {
	match := buildMatch(1, 10, 1, 2)
	mr := &matchSvcMockMatchRepo{
		matches: map[int64]*model.Match{1: match},
	}
	gr := &matchSvcMockGroupRepo{
		groups: map[int64]*model.Group{10: buildGroup(10, 5)},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr, hub: newTestHub()}

	err := svc.SetTableNumber(context.Background(), 1, 3, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.TableNumber == nil || *match.TableNumber != 3 {
		t.Errorf("expected table number 3, got %v", match.TableNumber)
	}
	if match.Status != model.MatchInProgress {
		t.Errorf("expected status IN_PROGRESS, got %s", match.Status)
	}
}

func TestSetTableNumber_TableAlreadyInUse(t *testing.T) {
	tableNum := 3
	match1 := &model.Match{
		MatchID:     1,
		GroupID:     10,
		Status:      model.MatchInProgress,
		TableNumber: &tableNum,
	}
	match2 := buildMatch(2, 10, 3, 4)
	mr := &matchSvcMockMatchRepo{
		matches: map[int64]*model.Match{1: match1, 2: match2},
	}
	gr := &matchSvcMockGroupRepo{
		groups: map[int64]*model.Group{10: buildGroup(10, 5)},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr, hub: newTestHub()}

	err := svc.SetTableNumber(context.Background(), 2, 3, 5)
	if err == nil {
		t.Fatal("expected error for table already in use")
	}
}

// --- RecalcGroupPoints (public) test ---

func TestRecalcGroupPoints_Public(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(0)
	matches := []model.Match{
		{MatchID: 1, GroupID: 10, GroupPlayer1ID: &gp1, GroupPlayer2ID: &gp2, Score1: &s1, Score2: &s2, Status: model.MatchDone},
	}
	players := buildGroupPlayers(10, []int64{1, 2})

	mr := &matchSvcMockMatchRepo{groupMatches: map[int64][]model.Match{10: matches}}
	gr := &matchSvcMockGroupRepo{
		groups:  map[int64]*model.Group{10: buildGroup(10, 5)},
		players: map[int64][]model.GroupPlayer{10: players},
	}
	svc := &matchService{matchRepo: mr, groupRepo: gr}

	err := svc.RecalcGroupPoints(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
