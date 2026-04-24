package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"league-api/internal/model"
	"league-api/internal/repository"
)

// ImportResult holds counts and row-level errors from a CSV import.
type ImportResult struct {
	Imported int
	Skipped  int
	Errors   []ImportError
}

// ImportError describes a per-row CSV import failure.
type ImportError struct {
	Row     int
	Message string
}

// PlayerProfile aggregates user info, rating history, and optional profile detail.
type PlayerProfile struct {
	model.User
	RatingHistory []model.RatingHistory       `json:"ratingHistory"`
	Profile       *model.PlayerProfileDetail  `json:"profile,omitempty"`
}

// PlayerService handles player CRUD and CSV import.
type PlayerService interface {
	CreatePlayer(ctx context.Context, firstName, lastName, email string) (*model.User, error)
	ImportCSV(ctx context.Context, r io.Reader) (ImportResult, error)
	GetProfile(ctx context.Context, userID int64) (*PlayerProfile, error)
	ListPlayers(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error)
	GetPlayerEvents(ctx context.Context, userID int64, limit, offset int) (*model.PlayerEventsPage, error)
}

type playerService struct {
	userRepo   repository.UserRepository
	ratingRepo repository.RatingRepository
	eventRepo  repository.EventRepository
	groupRepo  repository.GroupRepository
	matchRepo  repository.MatchRepository
	profileSvc ProfileService
}

func NewPlayerService(
	userRepo repository.UserRepository,
	ratingRepo repository.RatingRepository,
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	matchRepo repository.MatchRepository,
	profileSvc ProfileService,
) PlayerService {
	return &playerService{
		userRepo:   userRepo,
		ratingRepo: ratingRepo,
		eventRepo:  eventRepo,
		groupRepo:  groupRepo,
		matchRepo:  matchRepo,
		profileSvc: profileSvc,
	}
}

func (s *playerService) CreatePlayer(ctx context.Context, firstName, lastName, email string) (*model.User, error) {
	u := &model.User{
		FirstName:     firstName,
		LastName:      lastName,
		Email:         email,
		CurrentRating: 1500,
		Deviation:     350,
		Volatility:    0.06,
	}
	id, err := s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("playerService.CreatePlayer: %w", err)
	}
	return s.userRepo.GetByID(ctx, id)
}

func (s *playerService) ImportCSV(ctx context.Context, r io.Reader) (ImportResult, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return ImportResult{}, fmt.Errorf("read CSV header: %w", err)
	}
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}
	requiredCols := []string{"first_name", "last_name", "email"}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			return ImportResult{}, fmt.Errorf("missing required CSV column: %s", col)
		}
	}

	var result ImportResult
	rowNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Message: err.Error()})
			continue
		}

		firstName := record[colIndex["first_name"]]
		lastName := record[colIndex["last_name"]]
		email := record[colIndex["email"]]

		if firstName == "" || lastName == "" || email == "" {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Message: "first_name, last_name, and email are required"})
			continue
		}

		initialRating := 1500.0
		if idx, ok := colIndex["initial_rating"]; ok && idx < len(record) && record[idx] != "" {
			if v, err := strconv.ParseFloat(record[idx], 64); err == nil {
				initialRating = v
			}
		}

		// Check duplicate.
		existing, err := s.userRepo.GetByEmail(ctx, email)
		if err == nil && existing != nil {
			result.Skipped++
			continue
		}

		u := &model.User{
			FirstName:     firstName,
			LastName:      lastName,
			Email:         email,
			CurrentRating: initialRating,
			Deviation:     350,
			Volatility:    0.06,
		}
		if _, err := s.userRepo.Create(ctx, u); err != nil {
			result.Errors = append(result.Errors, ImportError{Row: rowNum, Message: fmt.Sprintf("insert: %v", err)})
			continue
		}
		result.Imported++
	}
	return result, nil
}

func (s *playerService) GetProfile(ctx context.Context, userID int64) (*PlayerProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("playerService.GetProfile: %w", err)
	}
	history, err := s.ratingRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("playerService.GetProfile rating: %w", err)
	}
	profile, _ := s.profileSvc.GetProfile(ctx, userID)
	return &PlayerProfile{User: *user, RatingHistory: history, Profile: profile}, nil
}

func (s *playerService) ListPlayers(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error) {
	if limit <= 0 {
		limit = 50
	}
	if q != "" {
		return s.userRepo.Search(ctx, q, limit, offset, sortBy)
	}
	return s.userRepo.List(ctx, limit, offset, sortBy)
}

func (s *playerService) GetPlayerEvents(ctx context.Context, userID int64, limit, offset int) (*model.PlayerEventsPage, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	events, total, err := s.eventRepo.ListEventsForPlayer(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("playerService.GetPlayerEvents: %w", err)
	}

	summaries := make([]model.PlayerEventSummary, 0, len(events))
	for _, ev := range events {
		delta, _ := s.ratingRepo.GetEventDeltaForUser(ctx, userID, ev.EventID)
		eventHistory, _ := s.ratingRepo.GetByUserInEvent(ctx, userID, ev.EventID)

		matchDelta := make(map[int64]float64, len(eventHistory))
		for _, rh := range eventHistory {
			matchDelta[rh.MatchID] = rh.Delta
		}

		var ratingBefore, ratingAfter *float64
		if len(eventHistory) > 0 {
			before := eventHistory[0].Rating - eventHistory[0].Delta
			after := eventHistory[len(eventHistory)-1].Rating
			ratingBefore = &before
			ratingAfter = &after
		}

		gpRecords, err := s.groupRepo.ListPlayerGroupsInEvent(ctx, userID, ev.EventID)
		if err != nil {
			return nil, err
		}

		groups := make([]model.PlayerGroupSummary, 0, len(gpRecords))
		for _, gp := range gpRecords {
			grp, err := s.groupRepo.GetByID(ctx, gp.GroupID)
			if err != nil {
				return nil, err
			}

			// Load all players in group to build gpID→userID map for opponent lookup.
			allPlayers, err := s.groupRepo.GetPlayers(ctx, gp.GroupID)
			if err != nil {
				return nil, err
			}
			gpToUser := make(map[int64]int64, len(allPlayers))
			for _, p := range allPlayers {
				gpToUser[p.GroupPlayerID] = p.UserID
			}

			allMatches, err := s.matchRepo.ListByGroup(ctx, gp.GroupID)
			if err != nil {
				return nil, err
			}

			var matchSummaries []model.PlayerMatchSummary
			for _, m := range allMatches {
				isP1 := m.GroupPlayer1ID != nil && *m.GroupPlayer1ID == gp.GroupPlayerID
				isP2 := m.GroupPlayer2ID != nil && *m.GroupPlayer2ID == gp.GroupPlayerID
				if !isP1 && !isP2 {
					continue
				}

				ms := model.PlayerMatchSummary{
					MatchID:  m.MatchID,
					Status:   m.Status,
					Withdraw: isP1 && m.Withdraw1 || isP2 && m.Withdraw2,
					OppWithdraw: isP1 && m.Withdraw2 || isP2 && m.Withdraw1,
				}
				if isP1 {
					ms.MyScore = m.Score1
					ms.OppScore = m.Score2
					if m.GroupPlayer2ID != nil {
						oppUID := gpToUser[*m.GroupPlayer2ID]
						ms.OpponentID = &oppUID
					}
				} else {
					ms.MyScore = m.Score2
					ms.OppScore = m.Score1
					if m.GroupPlayer1ID != nil {
						oppUID := gpToUser[*m.GroupPlayer1ID]
						ms.OpponentID = &oppUID
					}
				}

				if ms.OpponentID != nil {
					if opp, err := s.userRepo.GetByID(ctx, *ms.OpponentID); err == nil {
						ms.OpponentName = opp.FirstName + " " + opp.LastName
					}
				}

				if m.Status == model.MatchDone && ms.MyScore != nil && ms.OppScore != nil {
					won := *ms.MyScore > *ms.OppScore
					ms.Won = &won
				}

				if d, ok := matchDelta[m.MatchID]; ok {
					ms.RatingDelta = &d
				}
				matchSummaries = append(matchSummaries, ms)
			}

			groups = append(groups, model.PlayerGroupSummary{
				GroupID:  gp.GroupID,
				Division: grp.Division,
				GroupNo:  grp.GroupNo,
				Status:   grp.Status,
				Place:    gp.Place,
				Points:   gp.Points,
				Advances: gp.Advances,
				Recedes:  gp.Recedes,
				Matches:  matchSummaries,
			})
		}

		summaries = append(summaries, model.PlayerEventSummary{
			EventID:      ev.EventID,
			LeagueID:     ev.LeagueID,
			Title:        ev.Title,
			StartDate:    ev.StartDate,
			EndDate:      ev.EndDate,
			Status:       ev.Status,
			RatingDelta:  delta,
			RatingBefore: ratingBefore,
			RatingAfter:  ratingAfter,
			Groups:       groups,
		})
	}

	return &model.PlayerEventsPage{
		Events: summaries,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}, nil
}
