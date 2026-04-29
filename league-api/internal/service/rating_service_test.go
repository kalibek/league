package service

import (
	"context"
	"errors"
	"testing"

	"league-api/internal/model"
)

// ratingMockUserRepo is a user repo for rating service tests with error injection.
type ratingMockUserRepo struct {
	users         map[int64]*model.User
	updateErr     error
	resetErr      error
}

func (m *ratingMockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *ratingMockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}

func (m *ratingMockUserRepo) Create(ctx context.Context, u *model.User) (int64, error) { return 1, nil }

func (m *ratingMockUserRepo) List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *ratingMockUserRepo) Search(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error) {
	return nil, nil
}

func (m *ratingMockUserRepo) UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error {
	return m.updateErr
}

func (m *ratingMockUserRepo) ResetAllRatings(ctx context.Context) error {
	return m.resetErr
}

func (m *ratingMockUserRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	return nil
}

func (m *ratingMockUserRepo) UpdateName(ctx context.Context, userID int64, firstName, lastName string) error {
	return nil
}

// ratingMockGroupRepo supports GetPlayers and ListByEvent for rating service tests.
type ratingMockGroupRepo struct {
	players   map[int64][]model.GroupPlayer // groupID → players
	groups    map[int64][]model.Group       // eventID → groups
}

func (m *ratingMockGroupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	return &model.Group{GroupID: id}, nil
}

func (m *ratingMockGroupRepo) ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error) {
	return m.groups[eventID], nil
}

func (m *ratingMockGroupRepo) Create(ctx context.Context, g *model.Group) (int64, error) { return 1, nil }

func (m *ratingMockGroupRepo) UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error {
	return nil
}

func (m *ratingMockGroupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	return m.players[groupID], nil
}

func (m *ratingMockGroupRepo) GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *ratingMockGroupRepo) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	return 1, nil
}

func (m *ratingMockGroupRepo) UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error {
	return nil
}

func (m *ratingMockGroupRepo) RemovePlayer(ctx context.Context, groupPlayerID int64) error { return nil }

func (m *ratingMockGroupRepo) ResetGroupPlayers(ctx context.Context, groupID int64) error {
	return nil
}

func (m *ratingMockGroupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return nil, nil
}

func (m *ratingMockGroupRepo) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	return nil, nil
}

// ratingMockMatchRepo supports ListByGroup.
type ratingMockMatchRepo struct {
	matches map[int64][]model.Match // groupID → matches
}

func (m *ratingMockMatchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) {
	return nil, nil
}

func (m *ratingMockMatchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	return m.matches[groupID], nil
}

func (m *ratingMockMatchRepo) Create(ctx context.Context, match *model.Match) (int64, error) {
	return 1, nil
}

func (m *ratingMockMatchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error {
	return nil
}

func (m *ratingMockMatchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	return nil
}

func (m *ratingMockMatchRepo) BulkCreate(ctx context.Context, matches []model.Match) error {
	return nil
}

func (m *ratingMockMatchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error { return nil }

func (m *ratingMockMatchRepo) SetWithdraw(ctx context.Context, matchID int64, position int) error {
	return nil
}

// ratingMockRatingRepo supports InsertHistory, DeleteByGroup, DeleteAll.
type ratingMockRatingRepo struct {
	insertErr    error
	deleteErr    error
	deleteAllErr error
}

func (m *ratingMockRatingRepo) InsertHistory(ctx context.Context, rh *model.RatingHistory) error {
	return m.insertErr
}

func (m *ratingMockRatingRepo) GetByUser(ctx context.Context, userID int64) ([]model.RatingHistory, error) {
	return nil, nil
}

func (m *ratingMockRatingRepo) GetByUserInEvent(ctx context.Context, userID, eventID int64) ([]model.RatingHistory, error) {
	return nil, nil
}

func (m *ratingMockRatingRepo) DeleteByGroup(ctx context.Context, groupID int64) error {
	return m.deleteErr
}

func (m *ratingMockRatingRepo) DeleteAll(ctx context.Context) error {
	return m.deleteAllErr
}

func (m *ratingMockRatingRepo) GetEventDeltaForUser(ctx context.Context, userID, eventID int64) (float64, error) {
	return 0, nil
}

// ratingMockEventRepo returns a list of done events.
type ratingMockEventRepo struct {
	events    []model.LeagueEvent
	listErr   error
}

func (m *ratingMockEventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	for _, e := range m.events {
		if e.EventID == id {
			return &e, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *ratingMockEventRepo) ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	return nil, nil
}

func (m *ratingMockEventRepo) ListDone(ctx context.Context) ([]model.LeagueEvent, error) {
	return m.events, m.listErr
}

func (m *ratingMockEventRepo) Create(ctx context.Context, e *model.LeagueEvent) (int64, error) {
	return 1, nil
}

func (m *ratingMockEventRepo) UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error {
	return nil
}

func (m *ratingMockEventRepo) ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error) {
	return nil, 0, nil
}

// buildRatingSvc is a helper to construct a ratingService for tests.
func buildRatingSvc(ur *ratingMockUserRepo, gr *ratingMockGroupRepo, mr *ratingMockMatchRepo, rr *ratingMockRatingRepo, er *ratingMockEventRepo) *ratingService {
	return &ratingService{
		userRepo:   ur,
		groupRepo:  gr,
		matchRepo:  mr,
		ratingRepo: rr,
		eventRepo:  er,
	}
}

// --- NewRatingService test ---

func TestNewRatingService(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{}
	mr := &ratingMockMatchRepo{}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := NewRatingService(nil, ur, gr, mr, rr, er)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

// --- CalculateGroupRatings tests ---

func TestCalculateGroupRatings_NoPlayers(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{players: map[int64][]model.GroupPlayer{1: {}}}
	mr := &ratingMockMatchRepo{matches: map[int64][]model.Match{}}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalculateGroupRatings_WithNonCalculatedPlayer(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: 1, GroupID: 1, UserID: 10},
				{GroupPlayerID: 2, GroupID: 1, UserID: 20, IsNonCalculated: true},
			},
		},
	}
	mr := &ratingMockMatchRepo{matches: map[int64][]model.Match{1: {}}}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalculateGroupRatings_WithDoneMatchAndScores(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)

	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: gp1, GroupID: 1, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 1, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {
				{
					MatchID:        1,
					GroupID:        1,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Score1:         &s1,
					Score2:         &s2,
					Status:         model.MatchDone,
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalculateGroupRatings_Withdraw1(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)

	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: gp1, GroupID: 1, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 1, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {
				{
					MatchID:        1,
					GroupID:        1,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Withdraw1:      true,
					Status:         model.MatchDone,
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalculateGroupRatings_Withdraw2(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)

	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: gp1, GroupID: 1, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 1, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {
				{
					MatchID:        1,
					GroupID:        1,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Withdraw2:      true,
					Status:         model.MatchDone,
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalculateGroupRatings_SkipsDraftMatch(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)

	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: gp1, GroupID: 1, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 1, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {
				{
					MatchID:        1,
					GroupID:        1,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Score1:         &s1,
					Score2:         &s2,
					Status:         model.MatchDraft, // not done
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalculateGroupRatings_SkipsMatchWithNilScores(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)

	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: gp1, GroupID: 1, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 1, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {
				{
					MatchID:        1,
					GroupID:        1,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					// Score1 and Score2 are nil, Withdraw both false — data error, should skip
					Status: model.MatchDone,
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error for nil-score match: %v", err)
	}
}

func TestCalculateGroupRatings_InsertHistoryError(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)

	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: gp1, GroupID: 1, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 1, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {
				{
					MatchID:        1,
					GroupID:        1,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Score1:         &s1,
					Score2:         &s2,
					Status:         model.MatchDone,
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{insertErr: errors.New("insert error")}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from InsertHistory")
	}
}

func TestCalculateGroupRatings_UpdateUserRatingError(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)

	ur := &ratingMockUserRepo{
		users: map[int64]*model.User{
			10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
			20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		},
		updateErr: errors.New("update error"),
	}
	gr := &ratingMockGroupRepo{
		players: map[int64][]model.GroupPlayer{
			1: {
				{GroupPlayerID: gp1, GroupID: 1, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 1, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			1: {
				{
					MatchID:        1,
					GroupID:        1,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Score1:         &s1,
					Score2:         &s2,
					Status:         model.MatchDone,
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.CalculateGroupRatings(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from UpdateRating")
	}
}

// --- RecalculateGroupRatings tests ---

func TestRecalculateGroupRatings_Success(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{players: map[int64][]model.GroupPlayer{5: {}}}
	mr := &ratingMockMatchRepo{matches: map[int64][]model.Match{}}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.RecalculateGroupRatings(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecalculateGroupRatings_DeleteError(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{players: map[int64][]model.GroupPlayer{}}
	mr := &ratingMockMatchRepo{}
	rr := &ratingMockRatingRepo{deleteErr: errors.New("delete error")}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.RecalculateGroupRatings(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error from DeleteByGroup")
	}
}

// --- DeleteGroupRatings tests ---

func TestDeleteGroupRatings_Success(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{}
	mr := &ratingMockMatchRepo{}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.DeleteGroupRatings(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteGroupRatings_Error(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{}
	mr := &ratingMockMatchRepo{}
	rr := &ratingMockRatingRepo{deleteErr: errors.New("delete error")}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	err := svc.DeleteGroupRatings(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from DeleteByGroup")
	}
}

// --- RecalculateAllRatings tests ---

func TestRecalculateAllRatings_NoEvents(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{groups: map[int64][]model.Group{}}
	mr := &ratingMockMatchRepo{}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{events: []model.LeagueEvent{}}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	result, err := svc.RecalculateAllRatings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EventsProcessed != 0 {
		t.Errorf("expected 0 events processed, got %d", result.EventsProcessed)
	}
}

func TestRecalculateAllRatings_DeleteAllError(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{}
	mr := &ratingMockMatchRepo{}
	rr := &ratingMockRatingRepo{deleteAllErr: errors.New("delete all error")}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	_, err := svc.RecalculateAllRatings(context.Background())
	if err == nil {
		t.Fatal("expected error from DeleteAll")
	}
}

func TestRecalculateAllRatings_ResetUsersError(t *testing.T) {
	ur := &ratingMockUserRepo{
		users:    map[int64]*model.User{},
		resetErr: errors.New("reset error"),
	}
	gr := &ratingMockGroupRepo{}
	mr := &ratingMockMatchRepo{}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	_, err := svc.RecalculateAllRatings(context.Background())
	if err == nil {
		t.Fatal("expected error from ResetAllRatings")
	}
}

func TestRecalculateAllRatings_WithDoneEvent_AndDoneGroup(t *testing.T) {
	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)

	ur := &ratingMockUserRepo{users: map[int64]*model.User{
		10: {UserID: 10, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
		20: {UserID: 20, CurrentRating: 1500, Deviation: 350, Volatility: 0.06},
	}}
	gr := &ratingMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {{GroupID: 5, EventID: 1, Status: model.GroupDone}},
		},
		players: map[int64][]model.GroupPlayer{
			5: {
				{GroupPlayerID: gp1, GroupID: 5, UserID: 10},
				{GroupPlayerID: gp2, GroupID: 5, UserID: 20},
			},
		},
	}
	mr := &ratingMockMatchRepo{
		matches: map[int64][]model.Match{
			5: {
				{
					MatchID:        1,
					GroupID:        5,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Score1:         &s1,
					Score2:         &s2,
					Status:         model.MatchDone,
				},
			},
		},
	}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{
		events: []model.LeagueEvent{
			{EventID: 1, Status: model.EventDone},
		},
	}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	result, err := svc.RecalculateAllRatings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EventsProcessed != 1 {
		t.Errorf("expected 1 event processed, got %d", result.EventsProcessed)
	}
	if result.GroupsProcessed != 1 {
		t.Errorf("expected 1 group processed, got %d", result.GroupsProcessed)
	}
	if result.MatchesProcessed != 1 {
		t.Errorf("expected 1 match processed, got %d", result.MatchesProcessed)
	}
}

func TestRecalculateAllRatings_SkipsNonDoneGroups(t *testing.T) {
	ur := &ratingMockUserRepo{users: map[int64]*model.User{}}
	gr := &ratingMockGroupRepo{
		groups: map[int64][]model.Group{
			1: {{GroupID: 5, EventID: 1, Status: model.GroupInProgress}}, // not done
		},
		players: map[int64][]model.GroupPlayer{5: {}},
	}
	mr := &ratingMockMatchRepo{matches: map[int64][]model.Match{}}
	rr := &ratingMockRatingRepo{}
	er := &ratingMockEventRepo{
		events: []model.LeagueEvent{
			{EventID: 1, Status: model.EventDone},
		},
	}

	svc := buildRatingSvc(ur, gr, mr, rr, er)
	result, err := svc.RecalculateAllRatings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GroupsProcessed != 0 {
		t.Errorf("expected 0 groups processed (non-done skipped), got %d", result.GroupsProcessed)
	}
}
