package service

import (
	"context"
	"fmt"

	"league-api/internal/model"
	"league-api/internal/repository"
	"league-api/internal/ws"
)

// MatchService handles score entry, no-show, and placement recalculation.
type MatchService interface {
	UpdateScore(ctx context.Context, matchID int64, score1, score2 int16, gamesToWin int) error
	MarkNoShow(ctx context.Context, groupID, groupPlayerID int64, gamesToWin int) error
	RecalcGroupPoints(ctx context.Context, groupID int64) error
}

type matchService struct {
	matchRepo repository.MatchRepository
	groupRepo repository.GroupRepository
	hub       *ws.Hub
}

func NewMatchService(
	matchRepo repository.MatchRepository,
	groupRepo repository.GroupRepository,
	hub *ws.Hub,
) MatchService {
	return &matchService{
		matchRepo: matchRepo,
		groupRepo: groupRepo,
		hub:       hub,
	}
}

// UpdateScore validates and persists scores, recalculates group standings, and broadcasts.
func (s *matchService) UpdateScore(ctx context.Context, matchID int64, score1, score2 int16, gamesToWin int) error {
	m, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("matchService.UpdateScore get: %w", err)
	}

	// Validate: one score must equal gamesToWin, the other must be less.
	gtw := int16(gamesToWin)
	if !((score1 == gtw && score2 < gtw) || (score2 == gtw && score1 < gtw)) {
		return fmt.Errorf("invalid scores: one must equal %d and the other must be less", gamesToWin)
	}

	if err := s.matchRepo.UpdateScore(ctx, matchID, score1, score2); err != nil {
		return fmt.Errorf("matchService.UpdateScore: %w", err)
	}
	if err := s.matchRepo.UpdateStatus(ctx, matchID, model.MatchDone); err != nil {
		return fmt.Errorf("matchService.UpdateScore status: %w", err)
	}

	// Recalculate points for both players.
	if err := s.recalcGroupPoints(ctx, m.GroupID); err != nil {
		return fmt.Errorf("matchService.UpdateScore recalc: %w", err)
	}

	// Broadcast via WebSocket.
	s.hub.BroadcastToEvent(m.GroupID, ws.Message{
		Type:    "match_updated",
		GroupID: m.GroupID,
		MatchID: matchID,
		Payload: map[string]any{
			"matchId": matchID,
			"score1":  score1,
			"score2":  score2,
		},
	})

	return nil
}

// MarkNoShow marks all of a player's matches as withdrawn and recalculates standings.
func (s *matchService) MarkNoShow(ctx context.Context, groupID, groupPlayerID int64, gamesToWin int) error {
	matches, err := s.matchRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("matchService.MarkNoShow list: %w", err)
	}

	for _, m := range matches {
		if m.Status == model.MatchDone {
			continue
		}
		isP1 := m.GroupPlayer1ID != nil && *m.GroupPlayer1ID == groupPlayerID
		isP2 := m.GroupPlayer2ID != nil && *m.GroupPlayer2ID == groupPlayerID
		if !isP1 && !isP2 {
			continue
		}

		position := 2
		if isP1 {
			position = 1
		}

		// SetWithdraw marks the match DONE and sets the withdraw flag.
		if err := s.matchRepo.SetWithdraw(ctx, m.MatchID, position); err != nil {
			return fmt.Errorf("matchService.MarkNoShow withdraw: %w", err)
		}

		// Set scores: DNS player gets 0, opponent gets gamesToWin.
		var s1, s2 int16
		if isP1 {
			s1 = 0
			s2 = int16(gamesToWin)
		} else {
			s1 = int16(gamesToWin)
			s2 = 0
		}
		if err := s.matchRepo.UpdateScore(ctx, m.MatchID, s1, s2); err != nil {
			return fmt.Errorf("matchService.MarkNoShow score: %w", err)
		}
	}

	if err := s.recalcGroupPoints(ctx, groupID); err != nil {
		return fmt.Errorf("matchService.MarkNoShow recalc: %w", err)
	}

	s.hub.BroadcastToEvent(groupID, ws.Message{
		Type:    "match_updated",
		GroupID: groupID,
		Payload: map[string]any{"noShow": groupPlayerID},
	})

	return nil
}

func (s *matchService) RecalcGroupPoints(ctx context.Context, groupID int64) error {
	return s.recalcGroupPoints(ctx, groupID)
}

// recalcGroupPoints recomputes points and tiebreak points (game differential) for all players.
// Win = 2 pts, Loss = 1 pt; tiebreak = sum of (games won - games lost).
func (s *matchService) recalcGroupPoints(ctx context.Context, groupID int64) error {
	players, err := s.groupRepo.GetPlayers(ctx, groupID)
	if err != nil {
		return err
	}
	matches, err := s.matchRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return err
	}

	points := make(map[int64]int16)
	tiebreak := make(map[int64]int16)
	for _, p := range players {
		points[p.GroupPlayerID] = 0
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

		p1 := *m.GroupPlayer1ID
		p2 := *m.GroupPlayer2ID
		s1 := *m.Score1
		s2 := *m.Score2

		if m.Withdraw1 {
			points[p2] += 2
			tiebreak[p2] += s2
			tiebreak[p1] -= s2
		} else if m.Withdraw2 {
			points[p1] += 2
			tiebreak[p1] += s1
			tiebreak[p2] -= s1
		} else if s1 > s2 {
			points[p1] += 2
			points[p2] += 1
			tiebreak[p1] += s1 - s2
			tiebreak[p2] += s2 - s1
		} else {
			points[p2] += 2
			points[p1] += 1
			tiebreak[p2] += s2 - s1
			tiebreak[p1] += s1 - s2
		}
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
