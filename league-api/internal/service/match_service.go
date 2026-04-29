package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
	"league-api/internal/ws"
)

// MatchService handles score entry, no-show, and placement recalculation.
type MatchService interface {
	UpdateScore(ctx context.Context, matchID int64, score1, score2 int16, gamesToWin int, withdraw1, withdraw2 bool) error
	RecalcGroupPoints(ctx context.Context, groupID int64) error
}

type matchService struct {
	db        *sqlx.DB
	matchRepo repository.MatchRepository
	groupRepo repository.GroupRepository
	hub       *ws.Hub
}

func NewMatchService(
	db *sqlx.DB,
	matchRepo repository.MatchRepository,
	groupRepo repository.GroupRepository,
	hub *ws.Hub,
) MatchService {
	return &matchService{
		db:        db,
		matchRepo: matchRepo,
		groupRepo: groupRepo,
		hub:       hub,
	}
}

// UpdateScore validates and persists scores, recalculates group standings, and broadcasts.
func (s *matchService) UpdateScore(ctx context.Context, matchID int64, score1, score2 int16, gamesToWin int, withdraw1, withdraw2 bool) error {
	m, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("matchService.UpdateScore get: %w", err)
	}

	// Validate: one score must equal gamesToWin, the other must be less.
	gtw := int16(gamesToWin)
	if !((score1 == gtw && score2 < gtw) || (score2 == gtw && score1 < gtw)) {
		return fmt.Errorf("invalid scores: one must equal %d and the other must be less", gamesToWin)
	}

	if txErr := idb.RunInTx(ctx, s.db, func(txCtx context.Context) error {
		if err := s.matchRepo.UpdateScore(txCtx, matchID, score1, score2, withdraw1, withdraw2); err != nil {
			return fmt.Errorf("matchService.UpdateScore: %w", err)
		}
		if err := s.matchRepo.UpdateStatus(txCtx, matchID, model.MatchDone); err != nil {
			return fmt.Errorf("matchService.UpdateScore status: %w", err)
		}
		if err := s.recalcGroupPoints(txCtx, m.GroupID); err != nil {
			return fmt.Errorf("matchService.UpdateScore recalc: %w", err)
		}
		return nil
	}); txErr != nil {
		return txErr
	}

	// Broadcast via WebSocket — hub rooms are keyed by eventID, not groupID.
	grp, err := s.groupRepo.GetByID(ctx, m.GroupID)
	if err != nil {
		return fmt.Errorf("matchService.UpdateScore get group: %w", err)
	}
	s.hub.BroadcastToEvent(grp.EventID, ws.Message{
		Type:    "match_updated",
		GroupID: m.GroupID,
		MatchID: matchID,
		Payload: map[string]any{
			"matchId":   matchID,
			"score1":    score1,
			"score2":    score2,
			"withdraw1": withdraw1,
			"withdraw2": withdraw2,
		},
	})

	return nil
}

func (s *matchService) RecalcGroupPoints(ctx context.Context, groupID int64) error {
	return s.recalcGroupPoints(ctx, groupID)
}

// recalcGroupPoints recomputes points and tiebreak points for all players.
// Win = 2 pts, Loss = 1 pt.
// Tiebreak = score differential, but only from matches between players
// who share the same points total (tied group). Players with a unique
// points total get tiebreakPoints = 0.
func (s *matchService) recalcGroupPoints(ctx context.Context, groupID int64) error {
	players, err := s.groupRepo.GetPlayers(ctx, groupID)
	if err != nil {
		return err
	}
	matches, err := s.matchRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return err
	}

	// Pass 1: compute points for every player.
	points := make(map[int64]int16, len(players))
	for _, p := range players {
		points[p.GroupPlayerID] = 0
	}
	for _, m := range matches {
		if m.Status != model.MatchDone {
			continue
		}
		if m.GroupPlayer1ID == nil || m.GroupPlayer2ID == nil {
			continue
		}
		if m.Score1 == nil || m.Score2 == nil {
			continue
		}
		p1, p2 := *m.GroupPlayer1ID, *m.GroupPlayer2ID
		s1, s2 := *m.Score1, *m.Score2
		if m.Withdraw1 {
			points[p2] += 2
		} else if m.Withdraw2 {
			points[p1] += 2
		} else if s1 > s2 {
			points[p1] += 2
			points[p2] += 1
		} else {
			points[p2] += 2
			points[p1] += 1
		}
	}

	// Pass 2: compute tiebreak only within same-points groups.
	// Build sets of player IDs per points value (non-calculated players only).
	byPoints := make(map[int16]map[int64]bool)
	for _, p := range players {
		if p.IsNonCalculated {
			continue
		}
		pts := points[p.GroupPlayerID]
		if byPoints[pts] == nil {
			byPoints[pts] = make(map[int64]bool)
		}
		byPoints[pts][p.GroupPlayerID] = true
	}

	tiebreak := make(map[int64]int16, len(players))
	for _, p := range players {
		tiebreak[p.GroupPlayerID] = 0
	}
	for _, m := range matches {
		if m.Status != model.MatchDone {
			continue
		}
		if m.GroupPlayer1ID == nil || m.GroupPlayer2ID == nil {
			continue
		}
		if m.Score1 == nil || m.Score2 == nil {
			continue
		}
		if m.Withdraw1 || m.Withdraw2 {
			continue
		}
		p1, p2 := *m.GroupPlayer1ID, *m.GroupPlayer2ID
		s1, s2 := *m.Score1, *m.Score2
		// Only count if both players share the same points total.
		pts1 := points[p1]
		if byPoints[pts1] == nil || !byPoints[pts1][p2] {
			continue
		}
		tiebreak[p1] += s1 - s2
		tiebreak[p2] += s2 - s1
	}

	for i := range players {
		p := &players[i]
		p.Points = points[p.GroupPlayerID]
		p.TiebreakPoints = tiebreak[p.GroupPlayerID]
		if err := s.groupRepo.UpdatePlayer(ctx, p); err != nil {
			return err
		}
	}

	return nil
}
