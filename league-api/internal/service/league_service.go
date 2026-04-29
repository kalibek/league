package service

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

// LeagueRoleEntry is a role record enriched with user details.
type LeagueRoleEntry struct {
	UserID    int64  `json:"userId"`
	LeagueID  int64  `json:"leagueId"`
	RoleName  string `json:"roleName"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// LeagueService handles league CRUD and role management.
type LeagueService interface {
	CreateLeague(ctx context.Context, creatorID int64, title, description string, config model.LeagueConfig) (*model.League, error)
	GetLeague(ctx context.Context, leagueID int64) (*model.League, error)
	ListLeagues(ctx context.Context) ([]model.League, error)
	ListLeagueSummaries(ctx context.Context) ([]model.LeagueSummary, error)
	UpdateConfig(ctx context.Context, leagueID int64, config model.LeagueConfig) error
	AssignRole(ctx context.Context, leagueID, targetUserID int64, roleName string) error
	RemoveRole(ctx context.Context, leagueID, targetUserID int64, roleName string) error
	IsMaintainer(ctx context.Context, leagueID, userID int64) bool
	ListLeagueRoles(ctx context.Context, leagueID int64) ([]LeagueRoleEntry, error)
}

type leagueService struct {
	db         *sqlx.DB
	leagueRepo repository.LeagueRepository
	userRepo   repository.UserRepository
}

func NewLeagueService(db *sqlx.DB, leagueRepo repository.LeagueRepository, userRepo repository.UserRepository) LeagueService {
	return &leagueService{db: db, leagueRepo: leagueRepo, userRepo: userRepo}
}

// roleNameToID maps role names to the IDs seeded by migration 001.
func roleNameToID(name string) (int, error) {
	switch name {
	case "player":
		return 1, nil
	case "umpire":
		return 2, nil
	case "maintainer":
		return 3, nil
	default:
		return 0, fmt.Errorf("unknown role: %s", name)
	}
}

func (s *leagueService) CreateLeague(ctx context.Context, creatorID int64, title, description string, config model.LeagueConfig) (*model.League, error) {
	l := &model.League{
		Title:       title,
		Description: description,
		Config:      config,
	}
	var id int64
	maintainerID, _ := roleNameToID("maintainer")
	playerID, _ := roleNameToID("player")
	if txErr := idb.RunInTx(ctx, s.db, func(txCtx context.Context) error {
		var err error
		id, err = s.leagueRepo.Create(txCtx, l)
		if err != nil {
			return fmt.Errorf("leagueService.CreateLeague: %w", err)
		}
		if err := s.leagueRepo.AssignRole(txCtx, model.UserRole{
			UserID: creatorID, RoleID: maintainerID, LeagueID: id,
		}); err != nil {
			return fmt.Errorf("assign maintainer: %w", err)
		}
		if err := s.leagueRepo.AssignRole(txCtx, model.UserRole{
			UserID: creatorID, RoleID: playerID, LeagueID: id,
		}); err != nil {
			return fmt.Errorf("assign player: %w", err)
		}
		return nil
	}); txErr != nil {
		return nil, txErr
	}
	return s.leagueRepo.GetByID(ctx, id)
}

func (s *leagueService) GetLeague(ctx context.Context, leagueID int64) (*model.League, error) {
	return s.leagueRepo.GetByID(ctx, leagueID)
}

func (s *leagueService) ListLeagues(ctx context.Context) ([]model.League, error) {
	return s.leagueRepo.List(ctx)
}

func (s *leagueService) ListLeagueSummaries(ctx context.Context) ([]model.LeagueSummary, error) {
	stats, err := s.leagueRepo.ListWithStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("leagueService.ListLeagueSummaries: %w", err)
	}
	summaries := make([]model.LeagueSummary, 0, len(stats))
	for _, ls := range stats {
		maintainers, _ := s.leagueRepo.ListMaintainers(ctx, ls.LeagueID, 3)
		if maintainers == nil {
			maintainers = []model.LeagueMaintainer{}
		}
		summaries = append(summaries, model.LeagueSummary{
			LeagueWithStats: ls,
			Maintainers:     maintainers,
		})
	}
	return summaries, nil
}

func (s *leagueService) UpdateConfig(ctx context.Context, leagueID int64, config model.LeagueConfig) error {
	return s.leagueRepo.UpdateConfig(ctx, leagueID, config)
}

func (s *leagueService) AssignRole(ctx context.Context, leagueID, targetUserID int64, roleName string) error {
	roleID, err := roleNameToID(roleName)
	if err != nil {
		return err
	}
	return s.leagueRepo.AssignRole(ctx, model.UserRole{
		UserID: targetUserID, RoleID: roleID, LeagueID: leagueID,
	})
}

func (s *leagueService) RemoveRole(ctx context.Context, leagueID, targetUserID int64, roleName string) error {
	roleID, err := roleNameToID(roleName)
	if err != nil {
		return err
	}
	return s.leagueRepo.RemoveRole(ctx, targetUserID, leagueID, roleID)
}

func (s *leagueService) IsMaintainer(ctx context.Context, leagueID, userID int64) bool {
	urs, err := s.leagueRepo.GetUserRoles(ctx, userID, leagueID)
	if err != nil {
		return false
	}
	maintainerID, _ := roleNameToID("maintainer")
	for _, ur := range urs {
		if ur.RoleID == maintainerID {
			return true
		}
	}
	return false
}

func (s *leagueService) ListLeagueRoles(ctx context.Context, leagueID int64) ([]LeagueRoleEntry, error) {
	urs, err := s.leagueRepo.ListLeagueRoles(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("leagueService.ListLeagueRoles: %w", err)
	}
	result := make([]LeagueRoleEntry, 0, len(urs))
	for _, ur := range urs {
		entry := LeagueRoleEntry{
			UserID:   ur.UserID,
			LeagueID: ur.LeagueID,
			RoleName: roleIDToName(ur.RoleID),
		}
		if u, err := s.userRepo.GetByID(ctx, ur.UserID); err == nil {
			entry.FirstName = u.FirstName
			entry.LastName = u.LastName
			entry.Email = u.Email
		}
		result = append(result, entry)
	}
	return result, nil
}

func roleIDToName(id int) string {
	switch id {
	case 1:
		return "player"
	case 2:
		return "umpire"
	case 3:
		return "maintainer"
	default:
		return "unknown"
	}
}
