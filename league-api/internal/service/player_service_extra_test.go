package service

import (
	"context"
	"testing"

	"league-api/internal/model"
)

// playerTestGroupRepo wraps psGroupRepo and overrides ListPlayerGroupsInEvent.
type playerTestGroupRepo struct {
	*psGroupRepo
	gpMap map[int64][]model.GroupPlayer // userID → group player records
}

func (c *playerTestGroupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	return c.gpMap[userID], nil
}

// newPlayerTestGroupRepo builds the wrapped group repo with provided group data.
func newPlayerTestGroupRepo(base *psGroupRepo, gpMap map[int64][]model.GroupPlayer) *playerTestGroupRepo {
	return &playerTestGroupRepo{psGroupRepo: base, gpMap: gpMap}
}

// TestGetPlayerEvents_WithMatchesAsP1 covers the full inner-loop path for
// a player appearing as GroupPlayer1 in a finished match.
func TestGetPlayerEvents_WithMatchesAsP1(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice", LastName: "A"}
	ur.users[2] = &model.User{UserID: 2, FirstName: "Bob", LastName: "B"}

	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)
	delta := 12.5
	matchID := int64(100)

	// Use playerEventRatingRepo so GetByUserInEvent returns the per-match history
	// needed for building the matchDelta map and RatingDelta on matches.
	rr := newPlayerEventRatingRepo(map[int64][]model.RatingHistory{
		10: {
			{UserID: 1, MatchID: matchID, Delta: delta, Rating: 1512.5},
		},
	})

	events := []model.LeagueEvent{
		{EventID: 10, LeagueID: 1, Status: model.EventDone},
	}
	er := &psEventRepo{events: events, total: 1}

	groupRepo := newPSGroupRepo()
	groupRepo.groups[5] = []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 5, UserID: 1, Points: 2, Place: 1, Advances: true},
		{GroupPlayerID: gp2, GroupID: 5, UserID: 2, Points: 1, Place: 2},
	}
	groupRepo.groupDetail[5] = &model.Group{GroupID: 5, EventID: 10, Division: "A", GroupNo: 1, Status: model.GroupDone}

	cgr := newPlayerTestGroupRepo(groupRepo, map[int64][]model.GroupPlayer{
		1: {
			{GroupPlayerID: gp1, GroupID: 5, UserID: 1, Points: 2, Place: 1, Advances: true},
		},
	})

	matchRepo := &psMatchRepo{
		matches: map[int64][]model.Match{
			5: {
				{
					MatchID:        matchID,
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

	svc := &playerService{
		userRepo:   ur,
		ratingRepo: rr,
		eventRepo:  er,
		groupRepo:  cgr,
		matchRepo:  matchRepo,
		profileSvc: &nopProfileService{},
	}

	page, err := svc.GetPlayerEvents(context.Background(), 1, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(page.Events))
	}
	ev := page.Events[0]
	if len(ev.Groups) != 1 {
		t.Fatalf("expected 1 group summary, got %d", len(ev.Groups))
	}
	grpSummary := ev.Groups[0]
	if len(grpSummary.Matches) != 1 {
		t.Fatalf("expected 1 match summary, got %d", len(grpSummary.Matches))
	}
	ms := grpSummary.Matches[0]
	if ms.Won == nil || !*ms.Won {
		t.Errorf("expected Won=true for p1 winning 3:1")
	}
	// Verify rating delta is linked to match.
	if ms.RatingDelta == nil || *ms.RatingDelta != delta {
		t.Errorf("expected RatingDelta=%v, got %v", delta, ms.RatingDelta)
	}
}

// TestGetPlayerEvents_AsP2_Lose covers the player appearing as GroupPlayer2 and losing.
func TestGetPlayerEvents_AsP2_Lose(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice", LastName: "A"}
	ur.users[2] = &model.User{UserID: 2, FirstName: "Bob", LastName: "B"}

	gp1, gp2 := int64(1), int64(2)
	s1, s2 := int16(3), int16(1)

	rr := &psRatingRepo{history: map[int64][]model.RatingHistory{}}

	events := []model.LeagueEvent{
		{EventID: 10, LeagueID: 1, Status: model.EventDone},
	}
	er := &psEventRepo{events: events, total: 1}

	groupRepo := newPSGroupRepo()
	groupRepo.groups[5] = []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 5, UserID: 1, Points: 2, Place: 1},
		{GroupPlayerID: gp2, GroupID: 5, UserID: 2, Points: 1, Place: 2},
	}
	groupRepo.groupDetail[5] = &model.Group{GroupID: 5, EventID: 10, Division: "A", GroupNo: 1, Status: model.GroupDone}

	matchRepo := &psMatchRepo{
		matches: map[int64][]model.Match{
			5: {
				{
					MatchID:        200,
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

	// Player 2 is the viewer.
	cgr := newPlayerTestGroupRepo(groupRepo, map[int64][]model.GroupPlayer{
		2: {{GroupPlayerID: gp2, GroupID: 5, UserID: 2, Points: 1, Place: 2}},
	})

	svc := &playerService{
		userRepo:   ur,
		ratingRepo: rr,
		eventRepo:  er,
		groupRepo:  cgr,
		matchRepo:  matchRepo,
		profileSvc: &nopProfileService{},
	}

	page, err := svc.GetPlayerEvents(context.Background(), 2, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Events) != 1 {
		t.Fatalf("expected 1 event")
	}
	ev := page.Events[0]
	if len(ev.Groups) != 1 {
		t.Fatalf("expected 1 group")
	}
	ms := ev.Groups[0].Matches[0]
	if ms.Won == nil || *ms.Won {
		t.Errorf("expected Won=false for player 2 losing 1:3")
	}
}

// TestGetPlayerEvents_WithWithdrawal covers the withdraw/oppWithdraw paths.
func TestGetPlayerEvents_WithWithdrawal(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice", LastName: "A"}
	ur.users[2] = &model.User{UserID: 2, FirstName: "Bob", LastName: "B"}

	gp1, gp2 := int64(1), int64(2)

	rr := &psRatingRepo{history: map[int64][]model.RatingHistory{}}
	events := []model.LeagueEvent{
		{EventID: 10, LeagueID: 1, Status: model.EventDone},
	}
	er := &psEventRepo{events: events, total: 1}

	groupRepo := newPSGroupRepo()
	groupRepo.groups[5] = []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 5, UserID: 1},
		{GroupPlayerID: gp2, GroupID: 5, UserID: 2},
	}
	groupRepo.groupDetail[5] = &model.Group{GroupID: 5, EventID: 10, Status: model.GroupDone}

	matchRepo := &psMatchRepo{
		matches: map[int64][]model.Match{
			5: {
				{
					MatchID:        300,
					GroupID:        5,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Withdraw1:      true, // p1 withdrew
					Status:         model.MatchDone,
				},
			},
		},
	}

	// Test as player 1 (withdrawer).
	cgr := newPlayerTestGroupRepo(groupRepo, map[int64][]model.GroupPlayer{
		1: {{GroupPlayerID: gp1, GroupID: 5, UserID: 1}},
	})

	svc := &playerService{
		userRepo:   ur,
		ratingRepo: rr,
		eventRepo:  er,
		groupRepo:  cgr,
		matchRepo:  matchRepo,
		profileSvc: &nopProfileService{},
	}

	page, err := svc.GetPlayerEvents(context.Background(), 1, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ev := page.Events[0]
	ms := ev.Groups[0].Matches[0]
	if !ms.Withdraw {
		t.Errorf("expected Withdraw=true for player 1 who withdrew")
	}
	if ms.OppWithdraw {
		t.Errorf("expected OppWithdraw=false for player 1")
	}
}

// TestGetPlayerEvents_OppWithdrawal covers the opponent-withdraws path.
func TestGetPlayerEvents_OppWithdrawal(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1, FirstName: "Alice", LastName: "A"}
	ur.users[2] = &model.User{UserID: 2, FirstName: "Bob", LastName: "B"}

	gp1, gp2 := int64(1), int64(2)

	rr := &psRatingRepo{history: map[int64][]model.RatingHistory{}}
	events := []model.LeagueEvent{{EventID: 10, LeagueID: 1, Status: model.EventDone}}
	er := &psEventRepo{events: events, total: 1}

	groupRepo := newPSGroupRepo()
	groupRepo.groups[5] = []model.GroupPlayer{
		{GroupPlayerID: gp1, GroupID: 5, UserID: 1},
		{GroupPlayerID: gp2, GroupID: 5, UserID: 2},
	}
	groupRepo.groupDetail[5] = &model.Group{GroupID: 5, EventID: 10, Status: model.GroupDone}

	matchRepo := &psMatchRepo{
		matches: map[int64][]model.Match{
			5: {
				{
					MatchID:        400,
					GroupID:        5,
					GroupPlayer1ID: &gp1,
					GroupPlayer2ID: &gp2,
					Withdraw2:      true, // p2 withdrew; p1 is viewer
					Status:         model.MatchDone,
				},
			},
		},
	}

	cgr := newPlayerTestGroupRepo(groupRepo, map[int64][]model.GroupPlayer{
		1: {{GroupPlayerID: gp1, GroupID: 5, UserID: 1}},
	})

	svc := &playerService{
		userRepo:   ur,
		ratingRepo: rr,
		eventRepo:  er,
		groupRepo:  cgr,
		matchRepo:  matchRepo,
		profileSvc: &nopProfileService{},
	}

	page, err := svc.GetPlayerEvents(context.Background(), 1, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ms := page.Events[0].Groups[0].Matches[0]
	if ms.Withdraw {
		t.Errorf("expected Withdraw=false for player 1 (other player withdrew)")
	}
	if !ms.OppWithdraw {
		t.Errorf("expected OppWithdraw=true for player 1 (p2 withdrew)")
	}
}

// TestGetPlayerEvents_WithRatingHistory covers the ratingBefore/ratingAfter path.
func TestGetPlayerEvents_WithRatingHistory(t *testing.T) {
	ur := newPSUserRepo()
	ur.users[1] = &model.User{UserID: 1}

	// GetByUserInEvent is used to fetch eventHistory for ratingBefore/After.
	// We need a ratingRepo that returns history for GetByUserInEvent.
	type ratingRepoWithEventHistory struct {
		psRatingRepo
		eventHistory map[int64]map[int64][]model.RatingHistory // userID → eventID → history
	}

	// We'll override the service directly with a custom eventHistory mock.
	// Since psRatingRepo.GetByUserInEvent always returns nil, we use a different approach:
	// build a concrete ratingService mock that overrides the method.

	// The cleanest approach: build a custom RatingRepo that returns event history.
	type customRatingRepo struct {
		psRatingRepo
		evtHistory []model.RatingHistory
	}

	crr := &struct {
		psRatingRepo
		evtHistory []model.RatingHistory
	}{
		psRatingRepo: psRatingRepo{history: map[int64][]model.RatingHistory{}},
		evtHistory: []model.RatingHistory{
			{UserID: 1, MatchID: 1, Delta: 10.0, Rating: 1510.0},
			{UserID: 1, MatchID: 2, Delta: -5.0, Rating: 1505.0},
		},
	}
	_ = crr
	_ = ratingRepoWithEventHistory{}

	// Use a full mock that wraps psRatingRepo but overrides GetByUserInEvent.
	rr := newPlayerEventRatingRepo(map[int64][]model.RatingHistory{
		10: {
			{UserID: 1, MatchID: 1, Delta: 10.0, Rating: 1510.0},
			{UserID: 1, MatchID: 2, Delta: -5.0, Rating: 1505.0},
		},
	})

	events := []model.LeagueEvent{{EventID: 10, LeagueID: 1, Status: model.EventDone}}
	er := &psEventRepo{events: events, total: 1}

	groupRepo := newPSGroupRepo()
	groupRepo.groupDetail[5] = &model.Group{GroupID: 5, EventID: 10, Status: model.GroupDone}

	matchRepo := &psMatchRepo{matches: map[int64][]model.Match{5: {}}}

	cgr := newPlayerTestGroupRepo(groupRepo, map[int64][]model.GroupPlayer{})

	svc := &playerService{
		userRepo:   ur,
		ratingRepo: rr,
		eventRepo:  er,
		groupRepo:  cgr,
		matchRepo:  matchRepo,
		profileSvc: &nopProfileService{},
	}

	page, err := svc.GetPlayerEvents(context.Background(), 1, 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ev := page.Events[0]
	if ev.RatingBefore == nil || *ev.RatingBefore != 1500.0 { // 1510 - 10 = 1500
		t.Errorf("unexpected ratingBefore: %v", ev.RatingBefore)
	}
	if ev.RatingAfter == nil || *ev.RatingAfter != 1505.0 {
		t.Errorf("unexpected ratingAfter: %v", ev.RatingAfter)
	}
}

// playerEventRatingRepo overrides GetByUserInEvent for event-history tests.
type playerEventRatingRepo struct {
	psRatingRepo
	byEventID map[int64][]model.RatingHistory // eventID → history entries
}

func newPlayerEventRatingRepo(byEvent map[int64][]model.RatingHistory) *playerEventRatingRepo {
	return &playerEventRatingRepo{
		psRatingRepo: psRatingRepo{history: map[int64][]model.RatingHistory{}},
		byEventID:    byEvent,
	}
}

func (r *playerEventRatingRepo) GetByUserInEvent(ctx context.Context, userID, eventID int64) ([]model.RatingHistory, error) {
	return r.byEventID[eventID], nil
}
