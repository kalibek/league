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

// DuplicateGroup holds a set of users who share the same normalized name.
type DuplicateGroup struct {
	NormalizedName string       `json:"normalizedName"`
	Users          []model.User `json:"users"`
}

// MergeResult summarizes the outcome of a merge operation.
type MergeResult struct {
	TargetID        int64   `json:"targetId"`
	MergedSourceIDs []int64 `json:"mergedSourceIds"`
	ConflictGroups  []int64 `json:"conflictGroupIds"`
	RecalcFromEvent *int64  `json:"recalcFromEvent,omitempty"`
}

// PlayersPage holds a paginated list of players with total count.
type PlayersPage struct {
	Players []model.User `json:"players"`
	Total   int          `json:"total"`
}

// PlayerService handles player CRUD and CSV import.
type PlayerService interface {
	CreatePlayer(ctx context.Context, firstName, lastName, email string) (*model.User, error)
	ImportCSV(ctx context.Context, r io.Reader) (ImportResult, error)
	GetProfile(ctx context.Context, userID int64) (*PlayerProfile, error)
	ListPlayers(ctx context.Context, q string, limit, offset int, sortBy string) (*PlayersPage, error)
	GetPlayerEvents(ctx context.Context, userID int64, limit, offset int) (*model.PlayerEventsPage, error)
	FindDuplicates(ctx context.Context) ([]DuplicateGroup, error)
	MergeUsers(ctx context.Context, targetID int64, sourceIDs []int64) (*MergeResult, error)
}

type playerService struct {
	userRepo   repository.UserRepository
	ratingRepo repository.RatingRepository
	eventRepo  repository.EventRepository
	groupRepo  repository.GroupRepository
	matchRepo  repository.MatchRepository
	profileSvc ProfileService
	ratingSvc  RatingService
}

func NewPlayerService(
	userRepo repository.UserRepository,
	ratingRepo repository.RatingRepository,
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	matchRepo repository.MatchRepository,
	profileSvc ProfileService,
	ratingSvc RatingService,
) PlayerService {
	return &playerService{
		userRepo:   userRepo,
		ratingRepo: ratingRepo,
		eventRepo:  eventRepo,
		groupRepo:  groupRepo,
		matchRepo:  matchRepo,
		profileSvc: profileSvc,
		ratingSvc:  ratingSvc,
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

func (s *playerService) ListPlayers(ctx context.Context, q string, limit, offset int, sortBy string) (*PlayersPage, error) {
	// Clamp limit to [10, 100]; default to 25 if outside range.
	if limit < 10 || limit > 100 {
		limit = 25
	}

	var players []model.User
	var err error
	if q != "" {
		players, err = s.userRepo.Search(ctx, q, limit, offset, sortBy)
	} else {
		players, err = s.userRepo.List(ctx, limit, offset, sortBy)
	}
	if err != nil {
		return nil, fmt.Errorf("playerService.ListPlayers: %w", err)
	}

	total, err := s.userRepo.CountPlayers(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("playerService.ListPlayers count: %w", err)
	}

	return &PlayersPage{Players: players, Total: total}, nil
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

func (s *playerService) FindDuplicates(ctx context.Context) ([]DuplicateGroup, error) {
	users, err := s.userRepo.FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("playerService.FindDuplicates: %w", err)
	}

	byNorm := make(map[string][]model.User)
	for _, u := range users {
		key := normalizeName(u.LastName) + " " + normalizeName(u.FirstName)
		byNorm[key] = append(byNorm[key], u)
	}

	result := make([]DuplicateGroup, 0)
	for key, group := range byNorm {
		if len(group) > 1 {
			result = append(result, DuplicateGroup{NormalizedName: key, Users: group})
		}
	}
	return result, nil
}

func (s *playerService) MergeUsers(ctx context.Context, targetID int64, sourceIDs []int64) (*MergeResult, error) {
	res := &MergeResult{
		TargetID:        targetID,
		MergedSourceIDs: make([]int64, 0),
		ConflictGroups:  make([]int64, 0),
	}

	var minEventID int64
	hasMinEvent := false

	for _, sourceID := range sourceIDs {
		if sourceID == targetID {
			return nil, fmt.Errorf("playerService.MergeUsers: sourceID %d equals targetID", sourceID)
		}

		src, err := s.userRepo.GetByID(ctx, sourceID)
		if err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: source user %d not found: %w", sourceID, err)
		}
		if src.MergedIntoUserID != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: user %d already merged", sourceID)
		}

		// Track earliest event where source has rating history.
		evID, found, err := s.ratingRepo.GetEarliestEventIDForUser(ctx, sourceID)
		if err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: earliest event for %d: %w", sourceID, err)
		}
		if found {
			if !hasMinEvent || evID < minEventID {
				minEventID = evID
				hasMinEvent = true
			}
		}

		// Also check target's earliest event to recalc from the correct starting point.
		tevID, tFound, err := s.ratingRepo.GetEarliestEventIDForUser(ctx, targetID)
		if err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: earliest event for target %d: %w", targetID, err)
		}
		if tFound {
			if !hasMinEvent || tevID < minEventID {
				minEventID = tevID
				hasMinEvent = true
			}
		}

		// Find groups where both source and target are present.
		conflictIDs, err := s.groupRepo.FindConflictingGroupIDs(ctx, sourceID, targetID)
		if err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: conflict groups for %d: %w", sourceID, err)
		}
		for _, gid := range conflictIDs {
			res.ConflictGroups = append(res.ConflictGroups, gid)
		}

		// Mark source as DNS in conflict groups.
		for _, gid := range conflictIDs {
			if err := s.groupRepo.SetPlayerStatusByUser(ctx, gid, sourceID, model.PlayerStatusDNS); err != nil {
				return nil, fmt.Errorf("playerService.MergeUsers: set DNS group %d user %d: %w", gid, sourceID, err)
			}
		}

		// Transfer non-conflict group_players rows to target.
		if err := s.groupRepo.UpdateGroupPlayerUserID(ctx, sourceID, targetID, conflictIDs); err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: transfer group players %d→%d: %w", sourceID, targetID, err)
		}

		// Reassign rating_history rows to target.
		if err := s.userRepo.UpdateRatingHistory(ctx, sourceID, targetID); err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: transfer rating history %d→%d: %w", sourceID, targetID, err)
		}

		// Drop source profile (target's is kept).
		if err := s.groupRepo.DeletePlayerProfileByUser(ctx, sourceID); err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: delete profile %d: %w", sourceID, err)
		}

		// Soft-delete the source user.
		if err := s.userRepo.SoftDeleteMerged(ctx, sourceID, targetID); err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: soft delete %d: %w", sourceID, err)
		}

		res.MergedSourceIDs = append(res.MergedSourceIDs, sourceID)
	}

	// Recalculate ratings from the earliest affected event.
	if hasMinEvent {
		res.RecalcFromEvent = &minEventID
		if _, err := s.ratingSvc.RecalculateFromEvent(ctx, minEventID); err != nil {
			return nil, fmt.Errorf("playerService.MergeUsers: recalc from event %d: %w", minEventID, err)
		}
	}

	return res, nil
}
