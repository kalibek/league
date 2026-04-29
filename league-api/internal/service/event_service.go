package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

// EventService manages league event lifecycle.
type EventService interface {
	CreateDraftEvent(ctx context.Context, leagueID int64, title string, startDate, endDate time.Time) (*model.LeagueEvent, error)
	StartEvent(ctx context.Context, eventID int64) error
	ListEvents(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error)
	GetEvent(ctx context.Context, eventID int64) (*model.LeagueEvent, error)
	GetEventDetail(ctx context.Context, eventID int64) (*model.EventDetail, error)
}

type eventService struct {
	db        *sqlx.DB
	eventRepo repository.EventRepository
	groupRepo repository.GroupRepository
	matchRepo repository.MatchRepository
	userRepo  repository.UserRepository
}

func NewEventService(
	db *sqlx.DB,
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	matchRepo repository.MatchRepository,
	userRepo repository.UserRepository,
) EventService {
	return &eventService{db: db, eventRepo: eventRepo, groupRepo: groupRepo, matchRepo: matchRepo, userRepo: userRepo}
}

func (s *eventService) CreateDraftEvent(ctx context.Context, leagueID int64, title string, startDate, endDate time.Time) (*model.LeagueEvent, error) {
	events, err := s.eventRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("eventService.CreateDraftEvent list: %w", err)
	}
	for _, e := range events {
		if e.Status == model.EventDraft || e.Status == model.EventInProgress {
			return nil, fmt.Errorf("league %d already has an active event (id=%d, status=%s)", leagueID, e.EventID, e.Status)
		}
	}

	ev := &model.LeagueEvent{
		LeagueID:  leagueID,
		Status:    model.EventDraft,
		Title:     title,
		StartDate: startDate,
		EndDate:   endDate,
	}
	id, err := s.eventRepo.Create(ctx, ev)
	if err != nil {
		return nil, fmt.Errorf("eventService.CreateDraftEvent: %w", err)
	}
	return s.eventRepo.GetByID(ctx, id)
}

// StartEvent transitions event DRAFT → IN_PROGRESS and generates round-robin matches for all groups.
func (s *eventService) StartEvent(ctx context.Context, eventID int64) error {
	ev, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("eventService.StartEvent get: %w", err)
	}
	if ev.Status != model.EventDraft {
		return fmt.Errorf("event %d is not in DRAFT status (current: %s)", eventID, ev.Status)
	}

	groups, err := s.groupRepo.ListByEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("eventService.StartEvent list groups: %w", err)
	}

	// Pre-compute match stubs per group (reads only, outside transaction).
	type groupStubs struct {
		groupID int64
		stubs   []model.Match
	}
	allStubs := make([]groupStubs, 0, len(groups))
	for _, g := range groups {
		allPlayers, err := s.groupRepo.GetPlayers(ctx, g.GroupID)
		if err != nil {
			return fmt.Errorf("eventService.StartEvent get players for group %d: %w", g.GroupID, err)
		}
		var players []model.GroupPlayer
		for _, p := range allPlayers {
			if !p.IsNonCalculated {
				players = append(players, p)
			}
		}
		var stubs []model.Match
		for a := 0; a < len(players); a++ {
			for b := a + 1; b < len(players); b++ {
				p1 := players[a].GroupPlayerID
				p2 := players[b].GroupPlayerID
				stubs = append(stubs, model.Match{
					GroupID:        g.GroupID,
					GroupPlayer1ID: &p1,
					GroupPlayer2ID: &p2,
					Status:         model.MatchDraft,
				})
			}
		}
		allStubs = append(allStubs, groupStubs{groupID: g.GroupID, stubs: stubs})
	}

	return idb.RunInTx(ctx, s.db, func(txCtx context.Context) error {
		for _, gs := range allStubs {
			if len(gs.stubs) > 0 {
				if err := s.matchRepo.BulkCreate(txCtx, gs.stubs); err != nil {
					return fmt.Errorf("eventService.StartEvent bulk create matches for group %d: %w", gs.groupID, err)
				}
			}
			if err := s.groupRepo.UpdateStatus(txCtx, gs.groupID, model.GroupInProgress); err != nil {
				return fmt.Errorf("eventService.StartEvent update group %d status: %w", gs.groupID, err)
			}
		}
		return s.eventRepo.UpdateStatus(txCtx, eventID, model.EventInProgress)
	})
}

func (s *eventService) ListEvents(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	return s.eventRepo.ListByLeague(ctx, leagueID)
}

func (s *eventService) GetEvent(ctx context.Context, eventID int64) (*model.LeagueEvent, error) {
	return s.eventRepo.GetByID(ctx, eventID)
}

func (s *eventService) GetEventDetail(ctx context.Context, eventID int64) (*model.EventDetail, error) {
	ev, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetEventDetail get: %w", err)
	}
	groups, err := s.groupRepo.ListByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetEventDetail list groups: %w", err)
	}
	details := make([]model.GroupDetail, 0, len(groups))
	for _, g := range groups {
		players, err := s.groupRepo.GetPlayers(ctx, g.GroupID)
		if err != nil {
			return nil, fmt.Errorf("eventService.GetEventDetail players group %d: %w", g.GroupID, err)
		}
		for i := range players {
			if u, err := s.userRepo.GetByID(ctx, players[i].UserID); err == nil {
				players[i].User = u
			}
		}
		matches, err := s.matchRepo.ListByGroup(ctx, g.GroupID)
		if err != nil {
			return nil, fmt.Errorf("eventService.GetEventDetail matches group %d: %w", g.GroupID, err)
		}
		details = append(details, model.GroupDetail{Group: g, Players: players, Matches: matches})
	}
	return &model.EventDetail{LeagueEvent: *ev, Groups: details}, nil
}
