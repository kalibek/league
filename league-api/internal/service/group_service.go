package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"league-api/internal/model"
	"league-api/internal/repository"
)

// GroupService handles round-robin generation and placement calculation.
type GroupService interface {
	GenerateRoundRobin(ctx context.Context, groupID int64) error
	CalculatePlacements(ctx context.Context, groupID int64) (needsManual []int64, err error)
	SetManualPlace(ctx context.Context, groupPlayerID int64, place int16) error
	AddNonCalculatedPlayer(ctx context.Context, groupID, userID int64) error
	GetGroupDetail(ctx context.Context, groupID int64) (*model.Group, []model.GroupPlayer, []model.Match, error)
	ListGroups(ctx context.Context, eventID int64) ([]model.Group, error)
	CreateGroup(ctx context.Context, eventID int64, division string, groupNo int, scheduled time.Time) (*model.Group, error)
	SeedPlayer(ctx context.Context, groupID, userID int64) error
	RemovePlayer(ctx context.Context, groupPlayerID int64) error
}

type groupService struct {
	groupRepo repository.GroupRepository
	matchRepo repository.MatchRepository
	eventRepo repository.EventRepository
}

func NewGroupService(
	groupRepo repository.GroupRepository,
	matchRepo repository.MatchRepository,
	eventRepo repository.EventRepository,
) GroupService {
	return &groupService{
		groupRepo: groupRepo,
		matchRepo: matchRepo,
		eventRepo: eventRepo,
	}
}

func (s *groupService) GetGroupDetail(ctx context.Context, groupID int64) (*model.Group, []model.GroupPlayer, []model.Match, error) {
	grp, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("groupService.GetGroupDetail group: %w", err)
	}
	players, err := s.groupRepo.GetPlayers(ctx, groupID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("groupService.GetGroupDetail players: %w", err)
	}
	matches, err := s.matchRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("groupService.GetGroupDetail matches: %w", err)
	}
	return grp, players, matches, nil
}

// GenerateRoundRobin creates n*(n-1)/2 match stubs for the group.
// Non-calculated players are excluded from match generation.
func (s *groupService) GenerateRoundRobin(ctx context.Context, groupID int64) error {
	players, err := s.groupRepo.GetPlayers(ctx, groupID)
	if err != nil {
		return fmt.Errorf("groupService.GenerateRoundRobin: %w", err)
	}

	// Filter out non-calculated players.
	var calculated []model.GroupPlayer
	for _, p := range players {
		if !p.IsNonCalculated {
			calculated = append(calculated, p)
		}
	}

	var matches []model.Match
	for i := 0; i < len(calculated); i++ {
		for j := i + 1; j < len(calculated); j++ {
			p1ID := calculated[i].GroupPlayerID
			p2ID := calculated[j].GroupPlayerID
			matches = append(matches, model.Match{
				GroupID:        groupID,
				GroupPlayer1ID: &p1ID,
				GroupPlayer2ID: &p2ID,
				Status:         model.MatchDraft,
			})
		}
	}

	if len(matches) == 0 {
		return nil
	}
	return s.matchRepo.BulkCreate(ctx, matches)
}

// CalculatePlacements computes and persists placements for a group.
// Returns a list of groupPlayerIDs that require manual ordering (three-way tiebreak after tiebreak).
func (s *groupService) CalculatePlacements(ctx context.Context, groupID int64) ([]int64, error) {
	players, err := s.groupRepo.GetPlayers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("groupService.CalculatePlacements players: %w", err)
	}
	matches, err := s.matchRepo.ListByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("groupService.CalculatePlacements matches: %w", err)
	}

	// Exclude non-calculated players.
	var ranked []model.GroupPlayer
	for _, p := range players {
		if !p.IsNonCalculated {
			ranked = append(ranked, p)
		}
	}

	// Group ranked players by points to find tied groups.
	pointsGroups := make(map[int16][]model.GroupPlayer)
	for _, p := range ranked {
		pointsGroups[p.Points] = append(pointsGroups[p.Points], p)
	}

	// Calculate tiebreak_points only within tied groups (same points).
	// Players with unique points get tiebreakPoints = 0.
	tbPoints := make(map[int64]int16)
	for _, group := range pointsGroups {
		if len(group) < 2 {
			tbPoints[group[0].GroupPlayerID] = 0
			continue
		}
		tiedIDs := make(map[int64]bool, len(group))
		for _, p := range group {
			tiedIDs[p.GroupPlayerID] = true
		}
		for _, p := range group {
			tbPoints[p.GroupPlayerID] = computeTiebreakPoints(p.GroupPlayerID, tiedIDs, matches)
		}
	}

	// Update tiebreak_points in DB.
	for i := range ranked {
		ranked[i].TiebreakPoints = tbPoints[ranked[i].GroupPlayerID]
		if err := s.groupRepo.UpdatePlayer(ctx, &ranked[i]); err != nil {
			return nil, fmt.Errorf("groupService.CalculatePlacements update tiebreak: %w", err)
		}
	}

	// Sort by points DESC, then tiebreak DESC.
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Points != ranked[j].Points {
			return ranked[i].Points > ranked[j].Points
		}
		return ranked[i].TiebreakPoints > ranked[j].TiebreakPoints
	})

	var needsManual []int64
	currentPlace := int16(1)

	// Group players with equal points.
	i := 0
	for i < len(ranked) {
		// Find end of tie group by points.
		j := i + 1
		for j < len(ranked) && ranked[j].Points == ranked[i].Points {
			j++
		}
		tieGroup := ranked[i:j]

		if len(tieGroup) == 1 {
			tieGroup[0].Place = currentPlace
			if err := s.groupRepo.UpdatePlayer(ctx, &tieGroup[0]); err != nil {
				return nil, err
			}
			currentPlace++
		} else if len(tieGroup) == 2 {
			winner := headToHeadWinner(tieGroup[0].GroupPlayerID, tieGroup[1].GroupPlayerID, matches)
			if winner == tieGroup[0].GroupPlayerID {
				tieGroup[0].Place = currentPlace
				tieGroup[1].Place = currentPlace + 1
			} else if winner == tieGroup[1].GroupPlayerID {
				tieGroup[1].Place = currentPlace
				tieGroup[0].Place = currentPlace + 1
			} else {
				// True tie — use tiebreak points.
				if tieGroup[0].TiebreakPoints >= tieGroup[1].TiebreakPoints {
					tieGroup[0].Place = currentPlace
					tieGroup[1].Place = currentPlace + 1
				} else {
					tieGroup[1].Place = currentPlace
					tieGroup[0].Place = currentPlace + 1
				}
			}
			for k := range tieGroup {
				if err := s.groupRepo.UpdatePlayer(ctx, &tieGroup[k]); err != nil {
					return nil, err
				}
			}
			currentPlace += int16(len(tieGroup))
		} else {
			// Multiple tied players — sub-group by tiebreak points.
			subI := 0
			for subI < len(tieGroup) {
				subJ := subI + 1
				for subJ < len(tieGroup) && tieGroup[subJ].TiebreakPoints == tieGroup[subI].TiebreakPoints {
					subJ++
				}
				subGroup := tieGroup[subI:subJ]

				if len(subGroup) == 1 {
					subGroup[0].Place = currentPlace
					if err := s.groupRepo.UpdatePlayer(ctx, &subGroup[0]); err != nil {
						return nil, err
					}
					currentPlace++
				} else if len(subGroup) == 2 {
					winner := headToHeadWinner(subGroup[0].GroupPlayerID, subGroup[1].GroupPlayerID, matches)
					if winner == subGroup[0].GroupPlayerID {
						subGroup[0].Place = currentPlace
						subGroup[1].Place = currentPlace + 1
					} else if winner == subGroup[1].GroupPlayerID {
						subGroup[1].Place = currentPlace
						subGroup[0].Place = currentPlace + 1
					} else {
						subGroup[0].Place = currentPlace
						subGroup[1].Place = currentPlace + 1
					}
					for k := range subGroup {
						if err := s.groupRepo.UpdatePlayer(ctx, &subGroup[k]); err != nil {
							return nil, err
						}
					}
					currentPlace += int16(len(subGroup))
				} else {
					// Manual intervention required.
					for _, p := range subGroup {
						needsManual = append(needsManual, p.GroupPlayerID)
					}
					currentPlace += int16(len(subGroup))
				}
				subI = subJ
			}
		}
		i = j
	}

	return needsManual, nil
}

// SetManualPlace overrides the place for a group player.
// It uses a synthetic GroupPlayer record with only the fields needed by UpdatePlayer.
func (s *groupService) SetManualPlace(ctx context.Context, groupPlayerID int64, place int16) error {
	gp := &model.GroupPlayer{
		GroupPlayerID: groupPlayerID,
		Place:         place,
		// UpdatePlayer uses GroupPlayerID as the key; other fields will be patched but not hurt.
	}
	return s.groupRepo.UpdatePlayer(ctx, gp)
}

// AddNonCalculatedPlayer adds a replacement player to a group (excluded from rating+placement).
func (s *groupService) AddNonCalculatedPlayer(ctx context.Context, groupID, userID int64) error {
	gp := &model.GroupPlayer{
		GroupID:         groupID,
		UserID:          userID,
		IsNonCalculated: true,
	}
	_, err := s.groupRepo.AddPlayer(ctx, gp)
	return err
}


func (s *groupService) ListGroups(ctx context.Context, eventID int64) ([]model.Group, error) {
	return s.groupRepo.ListByEvent(ctx, eventID)
}

func (s *groupService) CreateGroup(ctx context.Context, eventID int64, division string, groupNo int, scheduled time.Time) (*model.Group, error) {
	ev, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("groupService.CreateGroup get event: %w", err)
	}
	if ev.Status != model.EventDraft {
		return nil, fmt.Errorf("cannot add groups to an event that is not in DRAFT status")
	}
	g := &model.Group{
		EventID:   eventID,
		Status:    model.GroupDraft,
		Division:  division,
		GroupNo:   groupNo,
		Scheduled: scheduled,
	}
	id, err := s.groupRepo.Create(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("groupService.CreateGroup: %w", err)
	}
	return s.groupRepo.GetByID(ctx, id)
}

func (s *groupService) SeedPlayer(ctx context.Context, groupID, userID int64) error {
	grp, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("groupService.SeedPlayer get group: %w", err)
	}
	ev, err := s.eventRepo.GetByID(ctx, grp.EventID)
	if err != nil {
		return fmt.Errorf("groupService.SeedPlayer get event: %w", err)
	}
	if ev.Status != model.EventDraft {
		return fmt.Errorf("cannot seed players into a non-DRAFT event")
	}

	// Check the player is not already in another group within this event (single query).
	existing, err := s.groupRepo.ListPlayerGroupsInEvent(ctx, userID, grp.EventID)
	if err != nil {
		return fmt.Errorf("groupService.SeedPlayer check existing: %w", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("player %d is already assigned to a group in this event", userID)
	}

	// Compute next seed number.
	currentPlayers, _ := s.groupRepo.GetPlayers(ctx, groupID)
	seed := int16(len(currentPlayers) + 1)

	gp := &model.GroupPlayer{
		GroupID:         groupID,
		UserID:          userID,
		Seed:            seed,
		IsNonCalculated: false,
	}
	_, err = s.groupRepo.AddPlayer(ctx, gp)
	return err
}

func (s *groupService) RemovePlayer(ctx context.Context, groupPlayerID int64) error {
	return s.groupRepo.RemovePlayer(ctx, groupPlayerID)
}

// computeTiebreakPoints calculates games won minus games lost for a player across all group matches.
func computeTiebreakPoints(groupPlayerID int64, tiedIDs map[int64]bool, matches []model.Match) int16 {
	var tb int16
	for _, m := range matches {
		if m.Status != model.MatchDone {
			continue
		}
		if m.Score1 == nil || m.Score2 == nil {
			continue
		}
		if m.GroupPlayer1ID != nil && *m.GroupPlayer1ID == groupPlayerID {
			if m.GroupPlayer2ID == nil || !tiedIDs[*m.GroupPlayer2ID] {
				continue
			}
			tb += *m.Score1 - *m.Score2
		} else if m.GroupPlayer2ID != nil && *m.GroupPlayer2ID == groupPlayerID {
			if m.GroupPlayer1ID == nil || !tiedIDs[*m.GroupPlayer1ID] {
				continue
			}
			tb += *m.Score2 - *m.Score1
		}
	}
	return tb
}

// headToHeadWinner returns the groupPlayerID of the winner of the head-to-head match,
// or 0 if no match found or it was a draw/withdraw scenario.
func headToHeadWinner(p1ID, p2ID int64, matches []model.Match) int64 {
	for _, m := range matches {
		if m.Status != model.MatchDone {
			continue
		}
		p1IsPlayer1 := m.GroupPlayer1ID != nil && *m.GroupPlayer1ID == p1ID &&
			m.GroupPlayer2ID != nil && *m.GroupPlayer2ID == p2ID
		p1IsPlayer2 := m.GroupPlayer1ID != nil && *m.GroupPlayer1ID == p2ID &&
			m.GroupPlayer2ID != nil && *m.GroupPlayer2ID == p1ID

		if p1IsPlayer1 {
			if m.Withdraw1 {
				return p2ID
			}
			if m.Withdraw2 {
				return p1ID
			}
			if m.Score1 != nil && m.Score2 != nil {
				if *m.Score1 > *m.Score2 {
					return p1ID
				}
				return p2ID
			}
		} else if p1IsPlayer2 {
			if m.Withdraw1 {
				return p1ID
			}
			if m.Withdraw2 {
				return p2ID
			}
			if m.Score1 != nil && m.Score2 != nil {
				if *m.Score1 > *m.Score2 {
					return p2ID
				}
				return p1ID
			}
		}
	}
	return 0
}
