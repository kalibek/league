package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
	"league-api/pkg/glicko2"
)

// RatingService orchestrates Glicko2 rating calculation for a group.
type RatingService interface {
	CalculateGroupRatings(ctx context.Context, groupID int64) error
	RecalculateGroupRatings(ctx context.Context, groupID int64) error
	DeleteGroupRatings(ctx context.Context, groupID int64) error
	RecalculateAllRatings(ctx context.Context) (RecalcResult, error)
}

// RecalcResult holds counts returned from a full recalculation.
type RecalcResult struct {
	EventsProcessed  int `json:"eventsProcessed"`
	GroupsProcessed  int `json:"groupsProcessed"`
	MatchesProcessed int `json:"matchesProcessed"`
}

type ratingService struct {
	db         *sqlx.DB
	userRepo   repository.UserRepository
	groupRepo  repository.GroupRepository
	matchRepo  repository.MatchRepository
	ratingRepo repository.RatingRepository
	eventRepo  repository.EventRepository
}

func NewRatingService(
	db *sqlx.DB,
	userRepo repository.UserRepository,
	groupRepo repository.GroupRepository,
	matchRepo repository.MatchRepository,
	ratingRepo repository.RatingRepository,
	eventRepo repository.EventRepository,
) RatingService {
	return &ratingService{
		db:         db,
		userRepo:   userRepo,
		groupRepo:  groupRepo,
		matchRepo:  matchRepo,
		ratingRepo: ratingRepo,
		eventRepo:  eventRepo,
	}
}

// CalculateGroupRatings runs sequential per-match Glicko2 for all calculated players in a group.
// Each DONE match is processed in match_id order; both players' ratings update simultaneously
// per match and carry forward to subsequent matches.
func (s *ratingService) CalculateGroupRatings(ctx context.Context, groupID int64) error {
	players, err := s.groupRepo.GetPlayers(ctx, groupID)
	if err != nil {
		return fmt.Errorf("ratingService.CalculateGroupRatings players: %w", err)
	}
	matches, err := s.matchRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("ratingService.CalculateGroupRatings matches: %w", err)
	}

	// Build state for calculated players only.
	gpToUserID := make(map[int64]int64)            // groupPlayerID → userID
	playerStates := make(map[int64]glicko2.Player) // groupPlayerID → Glicko2 state

	for _, p := range players {
		if p.IsNonCalculated {
			continue
		}
		u, err := s.userRepo.GetByID(ctx, p.UserID)
		if err != nil {
			return fmt.Errorf("ratingService load user %d: %w", p.UserID, err)
		}
		gpToUserID[p.GroupPlayerID] = p.UserID
		playerStates[p.GroupPlayerID] = glicko2.Player{
			Rating:     u.CurrentRating,
			Deviation:  u.Deviation,
			Volatility: u.Volatility,
		}
	}

	// Process each match sequentially (ListByGroup returns ORDER BY match_id ASC).
	for _, m := range matches {
		if m.Status != model.MatchDone {
			continue
		}
		if m.GroupPlayer1ID == nil || m.GroupPlayer2ID == nil {
			continue
		}

		gp1ID := *m.GroupPlayer1ID
		gp2ID := *m.GroupPlayer2ID

		p1State, ok1 := playerStates[gp1ID]
		p2State, ok2 := playerStates[gp2ID]
		if !ok1 || !ok2 {
			// At least one player is non-calculated; skip this match.
			continue
		}

		var score1, score2 float64
		switch {
		case m.Withdraw1:
			score1, score2 = 0.0, 1.0
		case m.Withdraw2:
			score1, score2 = 1.0, 0.0
		case m.Score1 != nil && m.Score2 != nil:
			if *m.Score1 > *m.Score2 {
				score1, score2 = 1.0, 0.0
			} else {
				score1, score2 = 0.0, 1.0
			}
		default:
			continue // DONE but no score and no withdraw — data error, skip
		}

		// Calculate new ratings simultaneously using pre-match states.
		newP1 := glicko2.Calculate(p1State, []glicko2.MatchResult{{Opponent: p2State, Score: score1}})
		newP2 := glicko2.Calculate(p2State, []glicko2.MatchResult{{Opponent: p1State, Score: score2}})

		delta1 := newP1.Rating - p1State.Rating
		delta2 := newP2.Rating - p2State.Rating

		if err := s.ratingRepo.InsertHistory(ctx, &model.RatingHistory{
			UserID:     gpToUserID[gp1ID],
			MatchID:    m.MatchID,
			Delta:      delta1,
			Rating:     newP1.Rating,
			Deviation:  newP1.Deviation,
			Volatility: newP1.Volatility,
		}); err != nil {
			return fmt.Errorf("ratingService insert history p1 match %d: %w", m.MatchID, err)
		}
		if err := s.ratingRepo.InsertHistory(ctx, &model.RatingHistory{
			UserID:     gpToUserID[gp2ID],
			MatchID:    m.MatchID,
			Delta:      delta2,
			Rating:     newP2.Rating,
			Deviation:  newP2.Deviation,
			Volatility: newP2.Volatility,
		}); err != nil {
			return fmt.Errorf("ratingService insert history p2 match %d: %w", m.MatchID, err)
		}

		// Carry updated states forward to subsequent matches.
		playerStates[gp1ID] = newP1
		playerStates[gp2ID] = newP2
	}

	// Persist final ratings to the users table.
	for gpID, state := range playerStates {
		userID := gpToUserID[gpID]
		if err := s.userRepo.UpdateRating(ctx, userID, state.Rating, state.Deviation, state.Volatility); err != nil {
			return fmt.Errorf("ratingService update user %d: %w", userID, err)
		}
	}

	return nil
}

// RecalculateGroupRatings deletes existing history for the group and recalculates.
func (s *ratingService) RecalculateGroupRatings(ctx context.Context, groupID int64) error {
	return idb.RunInTx(ctx, s.db, func(txCtx context.Context) error {
		if err := s.ratingRepo.DeleteByGroup(txCtx, groupID); err != nil {
			return fmt.Errorf("ratingService.RecalculateGroupRatings delete: %w", err)
		}
		return s.CalculateGroupRatings(txCtx, groupID)
	})
}

// RecalculateAllRatings wipes all rating history, resets every user to the initial
// Glicko2 parameters, then replays all DONE events in chronological order.
func (s *ratingService) RecalculateAllRatings(ctx context.Context) (RecalcResult, error) {
	var result RecalcResult

	// Read the list of done events before starting the transaction.
	events, err := s.eventRepo.ListDone(ctx)
	if err != nil {
		return result, fmt.Errorf("ratingService.RecalculateAllRatings list events: %w", err)
	}

	// Pre-fetch all group/match data outside the transaction (reads only).
	type groupData struct {
		group   model.Group
		matches []model.Match
	}
	type eventData struct {
		groups []groupData
	}
	eventDatas := make([]eventData, 0, len(events))
	for _, ev := range events {
		groups, err := s.groupRepo.ListByEvent(ctx, ev.EventID)
		if err != nil {
			return result, fmt.Errorf("ratingService.RecalculateAllRatings list groups event %d: %w", ev.EventID, err)
		}
		var gds []groupData
		for _, g := range groups {
			if g.Status != model.GroupDone {
				continue
			}
			matches, err := s.matchRepo.ListByGroup(ctx, g.GroupID)
			if err != nil {
				return result, fmt.Errorf("ratingService.RecalculateAllRatings list matches group %d: %w", g.GroupID, err)
			}
			gds = append(gds, groupData{group: g, matches: matches})
		}
		eventDatas = append(eventDatas, eventData{groups: gds})
	}

	if txErr := idb.RunInTx(ctx, s.db, func(txCtx context.Context) error {
		if err := s.ratingRepo.DeleteAll(txCtx); err != nil {
			return fmt.Errorf("ratingService.RecalculateAllRatings delete history: %w", err)
		}
		if err := s.userRepo.ResetAllRatings(txCtx); err != nil {
			return fmt.Errorf("ratingService.RecalculateAllRatings reset users: %w", err)
		}
		for i, ed := range eventDatas {
			for _, gd := range ed.groups {
				if err := s.CalculateGroupRatings(txCtx, gd.group.GroupID); err != nil {
					return fmt.Errorf("ratingService.RecalculateAllRatings calc group %d: %w", gd.group.GroupID, err)
				}
				result.GroupsProcessed++
				for _, m := range gd.matches {
					if m.Status == model.MatchDone {
						result.MatchesProcessed++
					}
				}
			}
			_ = i
			result.EventsProcessed++
		}
		return nil
	}); txErr != nil {
		return RecalcResult{}, txErr
	}

	return result, nil
}

func (s *ratingService) DeleteGroupRatings(ctx context.Context, groupID int64) error {
	if err := s.ratingRepo.DeleteByGroup(ctx, groupID); err != nil {
		return fmt.Errorf("ratingService.DeleteGroupRatings: %w", err)
	}
	return nil
}
