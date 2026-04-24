package service

import (
	"context"
	"testing"

	"league-api/internal/model"
)

// --- helpers ---

func int16p(v int16) *int16  { return &v }
func int64p(v int64) *int64  { return &v }

func makePlayers(ids []int64) []model.GroupPlayer {
	players := make([]model.GroupPlayer, len(ids))
	for i, id := range ids {
		players[i] = model.GroupPlayer{
			GroupPlayerID: id,
			GroupID:       1,
			UserID:        id,
			Seed:          int16(i + 1),
		}
	}
	return players
}

func doneMatch(p1, p2, s1, s2 int64) model.Match {
	return model.Match{
		MatchID:        p1*100 + p2,
		GroupID:        1,
		GroupPlayer1ID: int64p(p1),
		GroupPlayer2ID: int64p(p2),
		Score1:         int16p(int16(s1)),
		Score2:         int16p(int16(s2)),
		Status:         model.MatchDone,
	}
}

// --- computeTiebreakPoints ---

func TestComputeTiebreakPoints(t *testing.T) {
	matches := []model.Match{
		doneMatch(1, 2, 3, 1), // p1 wins 3:1
		doneMatch(1, 3, 3, 0), // p1 wins 3:0
	}

	tb := computeTiebreakPoints(1, matches)
	// (3-1) + (3-0) = 5
	if tb != 5 {
		t.Errorf("expected 5, got %d", tb)
	}

	tb2 := computeTiebreakPoints(2, matches)
	// (1-3) = -2
	if tb2 != -2 {
		t.Errorf("expected -2, got %d", tb2)
	}
}

// --- headToHeadWinner ---

func TestHeadToHeadWinner(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   int64
		matches  []model.Match
		expected int64
	}{
		{
			name:     "p1 wins",
			p1:       1,
			p2:       2,
			matches:  []model.Match{doneMatch(1, 2, 3, 1)},
			expected: 1,
		},
		{
			name:     "p2 wins",
			p1:       1,
			p2:       2,
			matches:  []model.Match{doneMatch(1, 2, 1, 3)},
			expected: 2,
		},
		{
			name:     "p2 wins (reversed match order)",
			p1:       1,
			p2:       2,
			matches:  []model.Match{doneMatch(2, 1, 3, 1)}, // p2 is player1 in match
			expected: 2,
		},
		{
			name:     "no match found",
			p1:       1,
			p2:       3,
			matches:  []model.Match{doneMatch(1, 2, 3, 1)},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := headToHeadWinner(tc.p1, tc.p2, tc.matches)
			if got != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, got)
			}
		})
	}
}

// --- CalculatePlacements via mock ---

// mockGroupRepo implements repository.GroupRepository for testing.
type mockGroupRepo struct {
	players map[int64][]model.GroupPlayer
	updated []model.GroupPlayer
}

func (m *mockGroupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	return &model.Group{GroupID: id, EventID: 1}, nil
}

func (m *mockGroupRepo) ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error) {
	return nil, nil
}

func (m *mockGroupRepo) Create(ctx context.Context, g *model.Group) (int64, error) {
	return 1, nil
}

func (m *mockGroupRepo) UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error {
	return nil
}

func (m *mockGroupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	return m.players[groupID], nil
}

func (m *mockGroupRepo) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	return 1, nil
}

func (m *mockGroupRepo) UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error {
	m.updated = append(m.updated, *gp)
	for i, p := range m.players[gp.GroupID] {
		if p.GroupPlayerID == gp.GroupPlayerID {
			m.players[gp.GroupID][i] = *gp
		}
	}
	return nil
}

func (m *mockGroupRepo) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }

func (m *mockGroupRepo) ResetGroupPlayers(ctx context.Context, groupID int64) error { return nil }

func (m *mockGroupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return nil, nil
}

type mockMatchRepo struct {
	matches map[int64][]model.Match
}

func (m *mockMatchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) {
	for _, matches := range m.matches {
		for _, match := range matches {
			if match.MatchID == id {
				return &match, nil
			}
		}
	}
	return nil, nil
}

func (m *mockMatchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	return m.matches[groupID], nil
}

func (m *mockMatchRepo) Create(ctx context.Context, match *model.Match) (int64, error) {
	return 1, nil
}

func (m *mockMatchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16) error {
	return nil
}

func (m *mockMatchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	return nil
}

func (m *mockMatchRepo) BulkCreate(ctx context.Context, matches []model.Match) error {
	return nil
}

func (m *mockMatchRepo) SetWithdraw(ctx context.Context, matchID int64, position int) error {
	return nil
}

func (m *mockMatchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error { return nil }

type mockEventRepo struct{}

func (m *mockEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	return &model.LeagueEvent{EventID: id}, nil
}

func (m *mockEventRepo) ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	return nil, nil
}

func (m *mockEventRepo) Create(ctx context.Context, e *model.LeagueEvent) (int64, error) {
	return 1, nil
}

func (m *mockEventRepo) UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error {
	return nil
}

func (m *mockEventRepo) ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error) {
	return nil, 0, nil
}

func TestCalculatePlacements_ClearWinner(t *testing.T) {
	// 3 players: p1 beats p2 and p3, p2 beats p3.
	players := makePlayers([]int64{1, 2, 3})
	players[0].Points = 4 // 2 wins
	players[1].Points = 3 // 1 win 1 loss
	players[2].Points = 2 // 2 losses

	matches := []model.Match{
		doneMatch(1, 2, 3, 1),
		doneMatch(1, 3, 3, 0),
		doneMatch(2, 3, 3, 1),
	}

	gr := &mockGroupRepo{
		players: map[int64][]model.GroupPlayer{1: players},
	}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: matches}}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	needsManual, err := svc.CalculatePlacements(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(needsManual) != 0 {
		t.Errorf("expected no manual placements, got %v", needsManual)
	}

	// Verify places.
	placeOf := func(id int64) int16 {
		for _, p := range gr.players[1] {
			if p.GroupPlayerID == id {
				return p.Place
			}
		}
		return 0
	}
	if placeOf(1) != 1 {
		t.Errorf("p1 should be place 1, got %d", placeOf(1))
	}
	if placeOf(2) != 2 {
		t.Errorf("p2 should be place 2, got %d", placeOf(2))
	}
	if placeOf(3) != 3 {
		t.Errorf("p3 should be place 3, got %d", placeOf(3))
	}
}

func TestCalculatePlacements_TwoWayTie_HeadToHead(t *testing.T) {
	// p1 and p2 tied on points; p1 beats p2 head-to-head.
	players := makePlayers([]int64{1, 2})
	players[0].Points = 3
	players[1].Points = 3

	matches := []model.Match{
		doneMatch(1, 2, 3, 1), // p1 wins
	}

	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: players}}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: matches}}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	needsManual, err := svc.CalculatePlacements(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(needsManual) != 0 {
		t.Errorf("expected no manual placements")
	}

	placeOf := func(id int64) int16 {
		for _, p := range gr.players[1] {
			if p.GroupPlayerID == id {
				return p.Place
			}
		}
		return 0
	}
	if placeOf(1) != 1 {
		t.Errorf("p1 should be 1st, got %d", placeOf(1))
	}
	if placeOf(2) != 2 {
		t.Errorf("p2 should be 2nd, got %d", placeOf(2))
	}
}
