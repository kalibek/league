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
	err := svc.SetPlayerStatus(context.Background(), 1, 1, model.PlayerStatusDNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	err := svc.SetPlayerStatus(context.Background(), 1, 42, model.PlayerStatusDNS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	err := svc.SetPlayerStatus(context.Background(), 1, 1, model.PlayerStatus("invalid"))
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
	err := svc.SetPlayerStatus(context.Background(), 1, 1, model.PlayerStatusDNS)
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
