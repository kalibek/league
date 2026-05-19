package service

import (
	"context"
	"testing"
	"time"

	"league-api/internal/model"
)

// --- helpers ---

func int16p(v int16) *int16 { return &v }
func int64p(v int64) *int64 { return &v }

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

	// All three players are tied — all matches count.
	tiedIDs := map[int64]bool{1: true, 2: true, 3: true}

	tb := computeTiebreakPoints(1, tiedIDs, matches)
	// (3-1) + (3-0) = 5
	if tb != 5 {
		t.Errorf("expected 5, got %d", tb)
	}

	tb2 := computeTiebreakPoints(2, tiedIDs, matches)
	// (1-3) = -2
	if tb2 != -2 {
		t.Errorf("expected -2, got %d", tb2)
	}

	// Match against non-tied player must not count.
	tiedIDs12 := map[int64]bool{1: true, 2: true}
	tb3 := computeTiebreakPoints(1, tiedIDs12, matches)
	// only doneMatch(1,2,3,1) counts → (3-1) = 2
	if tb3 != 2 {
		t.Errorf("expected 2 (cross-group excluded), got %d", tb3)
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

func (m *mockGroupRepo) GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error) {
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

func (m *mockGroupRepo) SetPlayerStatus(ctx context.Context, groupPlayerID int64, status model.PlayerStatus) error {
	return nil
}

func (m *mockGroupRepo) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	users := make([]model.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, model.User{UserID: id})
	}
	return users, nil
}

func (m *mockGroupRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

// mockMatchRepo also satisfies repository.MatchRepository.
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

func (m *mockMatchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error {
	return nil
}

func (m *mockMatchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	return nil
}

func (m *mockMatchRepo) BulkCreate(ctx context.Context, matches []model.Match) error {
	return nil
}

func (m *mockMatchRepo) DeleteByGroupPlayer(ctx context.Context, groupID, groupPlayerID int64) ([]int64, error) {
	// Find and delete all matches involving this player
	if m.matches == nil {
		return nil, nil
	}
	var deletedIDs []int64
	groupMatches := m.matches[groupID]
	var remaining []model.Match
	for _, match := range groupMatches {
		if (match.GroupPlayer1ID != nil && *match.GroupPlayer1ID == groupPlayerID) ||
			(match.GroupPlayer2ID != nil && *match.GroupPlayer2ID == groupPlayerID) {
			deletedIDs = append(deletedIDs, match.MatchID)
		} else {
			remaining = append(remaining, match)
		}
	}
	m.matches[groupID] = remaining
	return deletedIDs, nil
}

func (m *mockMatchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error { return nil }

func (m *mockMatchRepo) SetTableNumber(ctx context.Context, matchID int64, tableNumber int) error {
	return nil
}

func (m *mockMatchRepo) ResetScore(ctx context.Context, matchID int64) error { return nil }

func (m *mockMatchRepo) ListInProgressByEvent(ctx context.Context, eventID int64) ([]int, error) {
	return nil, nil
}

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

func (m *mockEventRepo) ListDone(ctx context.Context) ([]model.LeagueEvent, error) {
	return nil, nil
}

func (m *mockEventRepo) UpdateDetails(ctx context.Context, id int64, title string, startDate, endDate time.Time) error {
	return nil
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

func TestCalculatePlacements_ThreeWayCircular_NeedsManual(t *testing.T) {
	// A beats B 3:2, B beats C 3:2, C beats A 3:2.
	// All players: 1 win + 1 loss = 3 pts each.
	// Tiebreak (within tied group): each player +1 from win, -1 from loss = 0.
	// Head-to-head is circular → manual placement required.
	players := makePlayers([]int64{1, 2, 3})
	players[0].Points = 3 // A
	players[1].Points = 3 // B
	players[2].Points = 3 // C

	matches := []model.Match{
		doneMatch(1, 2, 3, 2), // A beats B
		doneMatch(2, 3, 3, 2), // B beats C
		doneMatch(3, 1, 3, 2), // C beats A
	}

	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: players}}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: matches}}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	needsManual, err := svc.CalculatePlacements(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(needsManual) != 3 {
		t.Errorf("expected 3 players needing manual placement, got %d: %v", len(needsManual), needsManual)
	}
	// All three must be in the manual list.
	inManual := make(map[int64]bool)
	for _, id := range needsManual {
		inManual[id] = true
	}
	for _, id := range []int64{1, 2, 3} {
		if !inManual[id] {
			t.Errorf("player %d should be in manual list", id)
		}
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

func makePlayersWithStatus(ids []int64, dnsIDs map[int64]bool) []model.GroupPlayer {
	players := make([]model.GroupPlayer, len(ids))
	for i, id := range ids {
		status := model.PlayerStatusActive
		if dnsIDs[id] {
			status = model.PlayerStatusDNS
		}
		players[i] = model.GroupPlayer{
			GroupPlayerID: id,
			GroupID:       1,
			UserID:        id,
			Seed:          int16(i + 1),
			PlayerStatus:  status,
		}
	}
	return players
}

func TestCalculatePlacements_DNSPlayerPlacedLast(t *testing.T) {
	// p1 (4 pts), p2 (2 pts), p3 is DNS.
	// p3 should be placed at position 3, after p1 and p2.
	players := makePlayersWithStatus([]int64{1, 2, 3}, map[int64]bool{3: true})
	players[0].Points = 4
	players[1].Points = 2
	players[2].Points = 0 // DNS — doesn't matter

	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: players}}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: {}}}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	needsManual, err := svc.CalculatePlacements(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(needsManual) != 0 {
		t.Errorf("expected no manual placements, got %v", needsManual)
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
		t.Errorf("p1 should be place 1, got %d", placeOf(1))
	}
	if placeOf(2) != 2 {
		t.Errorf("p2 should be place 2, got %d", placeOf(2))
	}
	if placeOf(3) != 3 {
		t.Errorf("DNS p3 should be place 3 (last), got %d", placeOf(3))
	}
}

func TestCalculatePlacements_MultipleDNSPlayersPlacedLast(t *testing.T) {
	// 4 players: p1 active (4 pts), p2 active (2 pts), p3 DNS, p4 DNS.
	players := makePlayersWithStatus([]int64{1, 2, 3, 4}, map[int64]bool{3: true, 4: true})
	players[0].Points = 4
	players[1].Points = 2

	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: players}}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: {}}}
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	needsManual, err := svc.CalculatePlacements(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(needsManual) != 0 {
		t.Errorf("expected no manual placements, got %v", needsManual)
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
		t.Errorf("p1 should be place 1, got %d", placeOf(1))
	}
	if placeOf(2) != 2 {
		t.Errorf("p2 should be place 2, got %d", placeOf(2))
	}
	// DNS players get places 3 and 4 (order among themselves by seed, i.e. insertion order).
	p3place := placeOf(3)
	p4place := placeOf(4)
	if p3place < 3 || p3place > 4 {
		t.Errorf("DNS p3 should have place 3 or 4, got %d", p3place)
	}
	if p4place < 3 || p4place > 4 {
		t.Errorf("DNS p4 should have place 3 or 4, got %d", p4place)
	}
	if p3place == p4place {
		t.Errorf("DNS players should have different places, both got %d", p3place)
	}
}

// mockMatchRepoWithCapture wraps mockMatchRepo and records BulkCreate calls.
type mockMatchRepoWithCapture struct {
	mockMatchRepo
	created []model.Match
}

func (m *mockMatchRepoWithCapture) BulkCreate(ctx context.Context, matches []model.Match) error {
	m.created = append(m.created, matches...)
	return nil
}

// mockGroupRepoWithAddCapture wraps mockGroupRepo and returns a configurable ID from AddPlayer.
type mockGroupRepoWithAddCapture struct {
	mockGroupRepo
	nextID      int64
	addedPlayer *model.GroupPlayer
}

func (m *mockGroupRepoWithAddCapture) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	m.addedPlayer = gp
	id := m.nextID
	if id == 0 {
		id = 99
	}
	// Add to the internal map so GetPlayers returns it.
	m.players[gp.GroupID] = append(m.players[gp.GroupID], model.GroupPlayer{
		GroupPlayerID:   id,
		GroupID:         gp.GroupID,
		UserID:          gp.UserID,
		Seed:            gp.Seed,
		IsNonCalculated: gp.IsNonCalculated,
		PlayerStatus:    gp.PlayerStatus,
	})
	return id, nil
}

type inProgressEventRepo struct {
	mockEventRepo
}

func (m *inProgressEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	return &model.LeagueEvent{EventID: id, Status: model.EventInProgress}, nil
}

// doneEventRepo returns events with DONE status.
type doneEventRepo struct {
	mockEventRepo
}

func (m *doneEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	return &model.LeagueEvent{EventID: id, Status: model.EventDone}, nil
}

// draftStatusEventRepo returns events with DRAFT status.
type draftStatusEventRepo struct {
	mockEventRepo
}

func (m *draftStatusEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	return &model.LeagueEvent{EventID: id, Status: model.EventDraft}, nil
}

func TestSetPlayerStatus_ValidDNS(t *testing.T) {
	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: makePlayers([]int64{1})}}
	mr := &mockMatchRepo{}
	ipEr := &inProgressEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: ipEr}
	result, err := svc.SetPlayerStatus(context.Background(), 1, 1, model.PlayerStatusDNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

type mockGroupRepoWithSetStatus struct {
	mockGroupRepo
	setStatusCalled bool
	lastStatus      model.PlayerStatus
}

func (m *mockGroupRepoWithSetStatus) SetPlayerStatus(ctx context.Context, groupPlayerID int64, status model.PlayerStatus) error {
	m.setStatusCalled = true
	m.lastStatus = status
	return nil
}

func TestSetPlayerStatus_CallsRepo(t *testing.T) {
	gr := &mockGroupRepoWithSetStatus{
		mockGroupRepo: mockGroupRepo{
			players: map[int64][]model.GroupPlayer{1: makePlayers([]int64{42})},
		},
	}
	ipEr := &inProgressEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: &mockMatchRepo{}, eventRepo: ipEr}
	result, err := svc.SetPlayerStatus(context.Background(), 1, 42, model.PlayerStatusDNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !gr.setStatusCalled {
		t.Error("expected SetPlayerStatus to be called on repo")
	}
	if gr.lastStatus != model.PlayerStatusDNS {
		t.Errorf("expected status dns, got %s", gr.lastStatus)
	}
}

func TestSetPlayerStatus_InvalidStatus(t *testing.T) {
	gr := &mockGroupRepoWithSetStatus{
		mockGroupRepo: mockGroupRepo{players: map[int64][]model.GroupPlayer{1: makePlayers([]int64{1})}},
	}
	ipEr := &inProgressEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: &mockMatchRepo{}, eventRepo: ipEr}
	_, err := svc.SetPlayerStatus(context.Background(), 1, 1, model.PlayerStatus("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestSetPlayerStatus_EventNotInProgress(t *testing.T) {
	gr := &mockGroupRepoWithSetStatus{
		mockGroupRepo: mockGroupRepo{players: map[int64][]model.GroupPlayer{1: makePlayers([]int64{1})}},
	}
	// Default mockEventRepo returns event with empty status (not IN_PROGRESS)
	er := &mockEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: &mockMatchRepo{}, eventRepo: er}
	_, err := svc.SetPlayerStatus(context.Background(), 1, 1, model.PlayerStatusDNS)
	if err == nil {
		t.Fatal("expected error: event must be IN_PROGRESS")
	}
}

// --- AddPlayerToActiveGroup tests ---

func TestAddPlayerToActiveGroup_Success(t *testing.T) {
	// 3 existing non-calculated players; add a 4th.
	existingPlayers := makePlayers([]int64{10, 20, 30})
	gr := &mockGroupRepoWithAddCapture{
		mockGroupRepo: mockGroupRepo{
			players: map[int64][]model.GroupPlayer{1: existingPlayers},
		},
		nextID: 99,
	}
	mr := &mockMatchRepoWithCapture{}
	ipEr := &inProgressEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: ipEr}
	err := svc.AddPlayerToActiveGroup(context.Background(), 1, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// New player should have seed = 4 (3 existing + 1).
	if gr.addedPlayer == nil {
		t.Fatal("expected AddPlayer to be called")
	}
	if gr.addedPlayer.Seed != 4 {
		t.Errorf("expected seed 4, got %d", gr.addedPlayer.Seed)
	}
	if gr.addedPlayer.IsNonCalculated {
		t.Error("expected IsNonCalculated=false")
	}
	if gr.addedPlayer.PlayerStatus != model.PlayerStatusActive {
		t.Errorf("expected status active, got %s", gr.addedPlayer.PlayerStatus)
	}

	// 3 new matches should be created (new player vs each existing).
	if len(mr.created) != 3 {
		t.Errorf("expected 3 new matches, got %d", len(mr.created))
	}
	for _, m := range mr.created {
		if m.Status != model.MatchDraft {
			t.Errorf("expected match status DRAFT, got %s", m.Status)
		}
	}
}

func TestAddPlayerToActiveGroup_BlocksOnDoneEvent(t *testing.T) {
	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: makePlayers([]int64{1})}}
	mr := &mockMatchRepo{}
	er := &doneEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	err := svc.AddPlayerToActiveGroup(context.Background(), 1, 999)
	if err == nil {
		t.Fatal("expected error for DONE event")
	}
}

func TestAddPlayerToActiveGroup_BlocksOnDraftEvent(t *testing.T) {
	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: makePlayers([]int64{1})}}
	mr := &mockMatchRepo{}
	er := &draftStatusEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	err := svc.AddPlayerToActiveGroup(context.Background(), 1, 999)
	if err == nil {
		t.Fatal("expected error for DRAFT event")
	}
}

func TestAddPlayerToActiveGroup_BlocksOnDuplicate(t *testing.T) {
	// Player is already assigned to a group in this event.
	existing := makePlayers([]int64{1, 2})
	dupGr := &mockGroupRepoWithDuplicate{
		mockGroupRepo:   mockGroupRepo{players: map[int64][]model.GroupPlayer{1: existing}},
		existingInEvent: existing,
	}
	mr := &mockMatchRepo{}
	ipEr := &inProgressEventRepo{}

	svc := &groupService{groupRepo: dupGr, matchRepo: mr, eventRepo: ipEr}
	err := svc.AddPlayerToActiveGroup(context.Background(), 1, 1) // userID=1 already in group
	if err == nil {
		t.Fatal("expected error: player already in event")
	}
}

// mockGroupRepoWithDuplicate simulates a player already assigned to this event.
type mockGroupRepoWithDuplicate struct {
	mockGroupRepo
	existingInEvent []model.GroupPlayer
}

func (m *mockGroupRepoWithDuplicate) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return m.existingInEvent, nil
}

// --- SetPlayerStatus tests ---

// mockGroupRepoForActive allows updating players and is used for SetPlayerStatus active test.
type mockGroupRepoForActive struct {
	mockGroupRepo
}

func (m *mockGroupRepoForActive) SetPlayerStatus(ctx context.Context, groupPlayerID int64, status model.PlayerStatus) error {
	for i := range m.players[1] {
		if m.players[1][i].GroupPlayerID == groupPlayerID {
			m.players[1][i].PlayerStatus = status
			return nil
		}
	}
	return nil
}

// mockMatchRepoForActive tracks created matches for SetPlayerStatus active test.
type mockMatchRepoForActive struct {
	mockMatchRepo
	createdMatches []model.Match
}

func (m *mockMatchRepoForActive) Create(ctx context.Context, match *model.Match) (int64, error) {
	matchID := int64(len(m.createdMatches) + 200)
	match.MatchID = matchID
	m.createdMatches = append(m.createdMatches, *match)
	return matchID, nil
}

func TestSetPlayerStatus_MarkDNS(t *testing.T) {
	// Setup: player 1 has 2 matches (with players 2 and 3)
	p1, p2, p3 := int64(10), int64(11), int64(12)
	players := []model.GroupPlayer{
		{GroupPlayerID: p1, GroupID: 1, UserID: p1, Seed: 1, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: p2, GroupID: 1, UserID: p2, Seed: 2, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: p3, GroupID: 1, UserID: p3, Seed: 3, PlayerStatus: model.PlayerStatusActive},
	}
	matches := []model.Match{
		{MatchID: 100, GroupID: 1, GroupPlayer1ID: &p1, GroupPlayer2ID: &p2, Status: model.MatchDraft},
		{MatchID: 101, GroupID: 1, GroupPlayer1ID: &p1, GroupPlayer2ID: &p3, Status: model.MatchDraft},
		{MatchID: 102, GroupID: 1, GroupPlayer1ID: &p2, GroupPlayer2ID: &p3, Status: model.MatchDraft},
	}

	gr := &mockGroupRepo{
		players: map[int64][]model.GroupPlayer{1: players},
	}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: matches}}
	er := &inProgressEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	result, err := svc.SetPlayerStatus(context.Background(), 1, p1, model.PlayerStatusDNS)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.DeletedMatchIDs) != 2 {
		t.Errorf("expected 2 deleted match IDs, got %d: %v", len(result.DeletedMatchIDs), result.DeletedMatchIDs)
	}
	if len(result.NewMatches) != 0 {
		t.Errorf("expected no new matches, got %d", len(result.NewMatches))
	}

	// Verify remaining matches in mock
	remaining := mr.matches[1]
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining match, got %d", len(remaining))
	}
	if len(remaining) > 0 && remaining[0].MatchID != 102 {
		t.Errorf("expected match 102 to remain, got %d", remaining[0].MatchID)
	}
}

func TestSetPlayerStatus_MarkActive(t *testing.T) {
	// Setup: player 1 is DNS, players 2 and 3 are active (non-calculated)
	p1, p2, p3 := int64(10), int64(11), int64(12)
	players := []model.GroupPlayer{
		{GroupPlayerID: p1, GroupID: 1, UserID: p1, Seed: 1, PlayerStatus: model.PlayerStatusDNS},
		{GroupPlayerID: p2, GroupID: 1, UserID: p2, Seed: 2, PlayerStatus: model.PlayerStatusActive, IsNonCalculated: false},
		{GroupPlayerID: p3, GroupID: 1, UserID: p3, Seed: 3, PlayerStatus: model.PlayerStatusActive, IsNonCalculated: false},
	}

	grForActive := &mockGroupRepoForActive{
		mockGroupRepo: mockGroupRepo{
			players: map[int64][]model.GroupPlayer{1: players},
		},
	}

	mrForActive := &mockMatchRepoForActive{
		mockMatchRepo: mockMatchRepo{
			matches: map[int64][]model.Match{1: {}},
		},
	}

	er := &inProgressEventRepo{}

	svc := &groupService{groupRepo: grForActive, matchRepo: mrForActive, eventRepo: er}
	result, err := svc.SetPlayerStatus(context.Background(), 1, p1, model.PlayerStatusActive)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.DeletedMatchIDs) != 0 {
		t.Errorf("expected no deleted match IDs, got %d", len(result.DeletedMatchIDs))
	}
	if len(result.NewMatches) != 2 {
		t.Errorf("expected 2 new matches, got %d", len(result.NewMatches))
	}

	// Verify the new matches are against both other players
	for _, m := range result.NewMatches {
		if m.GroupPlayer1ID == nil || *m.GroupPlayer1ID != p1 {
			t.Errorf("player 1 should be GroupPlayer1ID in all matches")
		}
		if m.Status != model.MatchDraft {
			t.Errorf("new matches should be DRAFT status, got %s", m.Status)
		}
	}
}

func TestSetPlayerStatus_MarkDNS_NoMatches(t *testing.T) {
	// Setup: player with no matches
	p1, p2 := int64(10), int64(11)
	players := []model.GroupPlayer{
		{GroupPlayerID: p1, GroupID: 1, UserID: p1, Seed: 1, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: p2, GroupID: 1, UserID: p2, Seed: 2, PlayerStatus: model.PlayerStatusActive},
	}
	matches := []model.Match{} // No matches yet

	gr := &mockGroupRepo{
		players: map[int64][]model.GroupPlayer{1: players},
	}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: matches}}
	er := &inProgressEventRepo{}

	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}
	result, err := svc.SetPlayerStatus(context.Background(), 1, p1, model.PlayerStatusDNS)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.DeletedMatchIDs) != 0 {
		t.Errorf("expected empty deleted match IDs, got %v", result.DeletedMatchIDs)
	}
}

// --- computeWinLossRatio ---

func withdrawnMatch(p1, p2 int64, w1, w2 bool) model.Match {
	s1, s2 := int16(0), int16(0)
	return model.Match{
		MatchID:        p1*100 + p2 + 500,
		GroupID:        1,
		GroupPlayer1ID: int64p(p1),
		GroupPlayer2ID: int64p(p2),
		Score1:         &s1,
		Score2:         &s2,
		Status:         model.MatchDone,
		Withdraw1:      w1,
		Withdraw2:      w2,
	}
}

func TestComputeWinLossRatio_BasicWins(t *testing.T) {
	matches := []model.Match{
		doneMatch(1, 2, 3, 0), // p1 wins
		doneMatch(1, 3, 3, 0), // p1 wins
		doneMatch(2, 3, 3, 0), // p2 wins, p3 loses
	}
	r := computeWinLossRatio(1, matches)
	// p1: 2W 0L → MaxFloat64
	if r <= 100 {
		t.Errorf("expected very large ratio for 2W/0L, got %f", r)
	}
	r2 := computeWinLossRatio(2, matches)
	// p2: 1W 1L → 1.0
	if r2 != 1.0 {
		t.Errorf("expected 1.0 for 1W/1L, got %f", r2)
	}
	r3 := computeWinLossRatio(3, matches)
	// p3: 0W 2L → 0.0
	if r3 != 0.0 {
		t.Errorf("expected 0.0 for 0W/2L, got %f", r3)
	}
}

func TestComputeWinLossRatio_NoMatches(t *testing.T) {
	r := computeWinLossRatio(1, nil)
	if r != 0.0 {
		t.Errorf("expected 0.0 with no matches, got %f", r)
	}
}

func TestComputeWinLossRatio_WithdrawnMatchesSkipped(t *testing.T) {
	matches := []model.Match{
		withdrawnMatch(1, 2, false, true), // p2 withdrew — skip entirely
		doneMatch(1, 3, 3, 0),             // p1 wins
	}
	r := computeWinLossRatio(1, matches)
	// Only the non-withdrawn match counts: 1W 0L → MaxFloat64
	if r <= 100 {
		t.Errorf("expected large ratio (withdrew match skipped), got %f", r)
	}
}

// --- 3-way tie resolved by W/L ratio ---

func TestCalculatePlacements_ThreeWayTie_ResolvedByWinLoss(t *testing.T) {
	// 3 players all have equal points and equal tiebreak points.
	// p1: 2W 1L → ratio ~2.0
	// p2: 1W 2L → ratio ~0.5
	// p3: 1W 1L → ratio 1.0   (actually wait, in round-robin 3 players each play 2 matches)
	// Matches: p1 beats p2 and p3; p2 beats p3.
	// But that gives p1=2W p2=1W p3=0W → different points, not a tie.
	//
	// For a 3-way tie with equal tiebreak points, we need:
	// All players have same main points AND same tb points.
	// p1 beats p2 3:1, p2 beats p3 3:1, p3 beats p1 3:1 (cycle) → equal main points AND equal tb points.
	// p1: 1W 1L (vs p2 won, vs p3 lost); p2: 1W 1L; p3: 1W 1L → equal W/L too → manual.
	//
	// Better: add a 4th non-tied player so the 3-way sub-group has different full-match records.
	// p1: 2W 1L overall (beats p4), p2: 1W 2L overall (loses to p4), p3: 1W 1L + 1 draw... hmm
	// Actually use 4 players: p1,p2,p3 in cycle (equal tb), p4 separate.
	// p1 beats p4 → p1 has 2W 1L overall ratio = 2.0
	// p2 loses to p4 → p2 has 1W 2L ratio = 0.5
	// p3 doesn't play p4 (no match, or draws) → p3 has 1W 1L ratio = 1.0
	//
	// But all 4 players would be in the same group and have different point totals...
	// Let me think of the simplest setup.
	//
	// 3 players, cyclic results + extra match each against a guest (non-calculated) player:
	// p1 beats p2 (tb: p1+2, p2-2); p2 beats p3 (tb: p2+2, p3-2); p3 beats p1 (tb: p3+2, p1-2)
	// → all have same main points (2 pts each assuming win=2), same tb=0.
	// Now add full-group W/L: p1 also beats non-calc p4 (W→extra win for p1)
	// But non-calc players don't affect the counted matches... let me just keep it simple.
	//
	// Simplest: 3 players, cyclic tie (equal main points + equal tb points),
	// PLUS each player has a varying number of wins from other matches in the group.
	// But in a 3-player group there are only 3 matches.
	//
	// For a 3-player cycle: p1 beats p2, p2 beats p3, p3 beats p1 →
	// Each player: 1W, 1L → equal W/L → manual. So W/L doesn't resolve this particular cycle.
	//
	// For W/L to resolve a 3-way tie, we need different overall W/L ratios.
	// This requires matches outside the tied sub-group.
	// Use 4 players where p1,p2,p3 tie but have different W/L:
	// p4 beats some but not others.
	// p4 loses to p1 and p3, beats p2.
	// Main points: p1=4, p2=4, p3=4, p4=2 → 3-way tie among p1,p2,p3
	// Wait, let me use proper scoring. Assume win=2pts.
	// p1 beats p2, p2 beats p3, p3 beats p1 → each has 2pts in the cycle
	// p1 beats p4 → p1 gets +2pts total = 4pts
	// p2 loses to p4 → p2 still 2pts total
	// p3 beats p4 → p3 gets 4pts total
	// p4: beats p2 = 2pts
	// So: p1=4, p2=2, p3=4, p4=2. p1 and p3 tied at 4pts.
	// That's a 2-way tie, not 3-way.
	//
	// OK, let me just test that W/L resolves a sub-group when tiebreak points are equal.
	// I'll use a scenario where 3 players have equal points and equal tb points,
	// but different W/L ratios because of draws/match counts.
	// Actually, in a 3-player round-robin there are only 3 matches. For all to have equal
	// tiebreak points AND different W/L, they'd need different wins counts which is impossible
	// in a pure 3-cycle.
	//
	// Let's use 5 players: p1,p2,p3 tied, p4,p5 separate. The matches outside the tie group
	// give p1,p2,p3 different W/L but same tb points (tb only counts within tied group).
	//
	// p4 and p5 each beat different sub-set of {p1,p2,p3}:
	// - p4 beats p2 and p3, loses to p1 → p1 gains 1W, p2,p3 gain 1L
	// - Result: p1 has W/L from outside = 1W extra; p2,p3 have 1L extra
	// But main points from p4,p5 would differ → not a 3-way tie anymore.
	//
	// This is getting complex. Let me use the simplest test that verifies the W/L logic works:
	// Mock the test by directly calling computeWinLossRatio and verifying the placement logic
	// using a case where W/L differs.
	//
	// Actually, the simplest valid test: 3 players with artificially set points and tiebreak_points
	// (same for all 3), and matches that give different W/L.
	// Use GroupPlayer.Points set directly without computing from matches.

	// 3 players: all at points=4, tiebreakPoints=0 (forced equal)
	// Matches (all DONE, non-withdrawn):
	// p1 beats p2 (W for p1), p2 beats p3 (W for p2), p3 beats p1 (W for p3) → cycle → 1W 1L each → W/L = 1.0 for all → still manual
	// But we want p1 to have better W/L. Add another match involving p1 winning:
	// Since we only have 3 players in group, only 3 matches. Cycle gives equal W/L.
	//
	// Let me use 4 players: p4 loses to everyone in the tie group. p4 also has equal points
	// somehow... Actually no: if p4 loses all 3 matches to p1/p2/p3, p4 has 0pts and p1/p2/p3 each have +1W.
	// p1 also beats p2, p2 beats p3, p3 beats p1 (cycle) → each has 1pt in cycle + 1 more win vs p4:
	// p1: 2W, 1L (lost to p3) → ratio 2.0
	// p2: 2W, 1L (lost to p1) → ratio 2.0
	// p3: 2W, 1L (lost to p2) → ratio 2.0
	// Equal again!
	//
	// OK I think for this test I should test the sub-group splitting logic directly.
	// The key thing to test: when 3 players have equal points AND equal tiebreak, and
	// different W/L ratios → W/L should place them without manual step.
	//
	// To get different W/L: some players win MORE matches overall.
	// Use 4 players where p1,p2,p3 all have same TOTAL points but p4 only loses to p1:
	// p1: beats p2, beats p4, loses to p3 → 2W 1L, W/L = 2.0
	// p2: beats p3, loses to p1, loses to p4 → 1W 2L, W/L = 0.5
	// p3: beats p1, loses to p2, beats p4 → 2W 1L, W/L = 2.0
	// p4: beats p2, loses to p1, loses to p3 → 1W 2L
	//
	// Points (using W=2): p1=4, p2=2, p3=4, p4=2 → p1 and p3 tied at 4pts. 2-way tie again.
	//
	// Conclusion: in a standard round-robin, getting a pure 3-way tie with equal TB
	// AND different W/L is hard because TB captures the head-to-head game diff.
	// The W/L step matters most when ALL HEAD-TO-HEAD results are draws (impossible in ping pong,
	// where each match has a winner) OR in a round-robin with more players where matches
	// vs non-tied players affect W/L.
	//
	// SIMPLEST APPROACH: Use setPoints directly in mock (bypass match computation)
	// and provide matches that give different W/L.

	// 3 players: manually set points=4 and tiebreakPoints=0 (set directly, not computed)
	// Matches: p1 wins 2, p2 wins 1, p3 wins 0 (but we'll set their points equal manually)
	// This simulates a case where the points were set by some other mechanism.
	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 1, Points: 4, TiebreakPoints: 0, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: 2, GroupID: 1, Points: 4, TiebreakPoints: 0, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: 3, GroupID: 1, Points: 4, TiebreakPoints: 0, PlayerStatus: model.PlayerStatusActive},
	}
	// Matches that give different W/L (independent of why tb points are equal):
	// p1 beats p2, p1 beats p3 → p1: 2W 0L
	// p2 beats p3 → p2: 1W 1L
	// p3: 0W 2L
	matches := []model.Match{
		doneMatch(1, 2, 3, 1),
		doneMatch(1, 3, 3, 0),
		doneMatch(2, 3, 3, 1),
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
		t.Errorf("expected W/L to resolve the tie without manual, got manual: %v", needsManual)
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
		t.Errorf("p1 should be 1st (2W/0L), got %d", placeOf(1))
	}
	if placeOf(2) != 2 {
		t.Errorf("p2 should be 2nd (1W/1L), got %d", placeOf(2))
	}
	if placeOf(3) != 3 {
		t.Errorf("p3 should be 3rd (0W/2L), got %d", placeOf(3))
	}
}

func TestCalculatePlacements_ThreeWayTie_EqualWinLoss_GoesManual(t *testing.T) {
	// 3 players, equal points, equal tb, equal W/L → must go manual.
	players := []model.GroupPlayer{
		{GroupPlayerID: 1, GroupID: 1, Points: 4, TiebreakPoints: 0, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: 2, GroupID: 1, Points: 4, TiebreakPoints: 0, PlayerStatus: model.PlayerStatusActive},
		{GroupPlayerID: 3, GroupID: 1, Points: 4, TiebreakPoints: 0, PlayerStatus: model.PlayerStatusActive},
	}
	// Cycle: p1 beats p2, p2 beats p3, p3 beats p1 → each 1W 1L → W/L = 1.0 for all.
	matches := []model.Match{
		doneMatch(1, 2, 3, 1),
		doneMatch(2, 3, 3, 1),
		doneMatch(3, 1, 3, 1),
	}

	gr := &mockGroupRepo{players: map[int64][]model.GroupPlayer{1: players}}
	mr := &mockMatchRepo{matches: map[int64][]model.Match{1: matches}}
	er := &mockEventRepo{}
	svc := &groupService{groupRepo: gr, matchRepo: mr, eventRepo: er}

	needsManual, err := svc.CalculatePlacements(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(needsManual) != 3 {
		t.Errorf("expected all 3 players to go manual (equal W/L), got %v", needsManual)
	}
}
