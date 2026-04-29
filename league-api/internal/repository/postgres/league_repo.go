package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

type leagueRepo struct {
	pool *sqlx.DB
}

func NewLeagueRepo(db *sqlx.DB) repository.LeagueRepository {
	return &leagueRepo{pool: db}
}

func (r *leagueRepo) db(ctx context.Context) idb.DBTX {
	if tx := idb.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *leagueRepo) GetByID(ctx context.Context, id int64) (*model.League, error) {
	var l model.League
	err := r.db(ctx).GetContext(ctx, &l, `SELECT * FROM leagues WHERE league_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("leagueRepo.GetByID: %w", err)
	}
	return &l, nil
}

func (r *leagueRepo) List(ctx context.Context) ([]model.League, error) {
	leagues := make([]model.League, 0)
	if err := r.db(ctx).SelectContext(ctx, &leagues, `SELECT * FROM leagues ORDER BY created DESC`); err != nil {
		return nil, fmt.Errorf("leagueRepo.List: %w", err)
	}
	return leagues, nil
}

func (r *leagueRepo) ListWithStats(ctx context.Context) ([]model.LeagueWithStats, error) {
	out := make([]model.LeagueWithStats, 0)
	const q = `
		SELECT
			l.league_id, l.title, l.description, l.configuration, l.created, l.last_updated,
			COUNT(le.event_id) AS event_count,
			MAX(le.end_date)   AS latest_event_date
		FROM leagues l
		LEFT JOIN league_events le ON le.league_id = l.league_id
		GROUP BY l.league_id
		ORDER BY MAX(le.end_date) DESC NULLS LAST, l.created DESC`
	if err := r.db(ctx).SelectContext(ctx, &out, q); err != nil {
		return nil, fmt.Errorf("leagueRepo.ListWithStats: %w", err)
	}
	return out, nil
}

func (r *leagueRepo) ListMaintainers(ctx context.Context, leagueID int64, limit int) ([]model.LeagueMaintainer, error) {
	out := make([]model.LeagueMaintainer, 0)
	const q = `
		SELECT u.user_id, u.first_name, u.last_name
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.user_id
		WHERE ur.league_id = $1 AND ur.role_id = 3
		ORDER BY u.last_name, u.first_name
		LIMIT $2`
	if err := r.db(ctx).SelectContext(ctx, &out, q, leagueID, limit); err != nil {
		return nil, fmt.Errorf("leagueRepo.ListMaintainers: %w", err)
	}
	return out, nil
}

func (r *leagueRepo) Create(ctx context.Context, l *model.League) (int64, error) {
	const q = `
		INSERT INTO leagues (title, description, configuration)
		VALUES ($1, $2, $3)
		RETURNING league_id`
	var id int64
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx, q, l.Title, l.Description, l.Config).Scan(&id)
	} else {
		err = r.pool.QueryRowContext(ctx, q, l.Title, l.Description, l.Config).Scan(&id)
	}
	if err != nil {
		return 0, fmt.Errorf("leagueRepo.Create: %w", err)
	}
	return id, nil
}

func (r *leagueRepo) UpdateConfig(ctx context.Context, id int64, config model.LeagueConfig) error {
	const q = `UPDATE leagues SET configuration = $1, last_updated = NOW() WHERE league_id = $2`
	_, err := r.db(ctx).ExecContext(ctx, q, config, id)
	if err != nil {
		return fmt.Errorf("leagueRepo.UpdateConfig: %w", err)
	}
	return nil
}

func (r *leagueRepo) AssignRole(ctx context.Context, ur model.UserRole) error {
	const q = `
		INSERT INTO user_roles (user_id, role_id, league_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`
	_, err := r.db(ctx).ExecContext(ctx, q, ur.UserID, ur.RoleID, ur.LeagueID)
	if err != nil {
		return fmt.Errorf("leagueRepo.AssignRole: %w", err)
	}
	return nil
}

func (r *leagueRepo) RemoveRole(ctx context.Context, userID, leagueID int64, roleID int) error {
	const q = `DELETE FROM user_roles WHERE user_id = $1 AND league_id = $2 AND role_id = $3`
	_, err := r.db(ctx).ExecContext(ctx, q, userID, leagueID, roleID)
	if err != nil {
		return fmt.Errorf("leagueRepo.RemoveRole: %w", err)
	}
	return nil
}

func (r *leagueRepo) GetUserRoles(ctx context.Context, userID, leagueID int64) ([]model.UserRole, error) {
	roles := make([]model.UserRole, 0)
	err := r.db(ctx).SelectContext(ctx, &roles,
		`SELECT user_id, role_id, league_id FROM user_roles WHERE user_id = $1 AND league_id = $2`,
		userID, leagueID,
	)
	if err != nil {
		return nil, fmt.Errorf("leagueRepo.GetUserRoles: %w", err)
	}
	return roles, nil
}

func (r *leagueRepo) GetAllUserRoles(ctx context.Context, userID int64) ([]model.UserRole, error) {
	roles := make([]model.UserRole, 0)
	err := r.db(ctx).SelectContext(ctx, &roles,
		`SELECT user_id, role_id, league_id FROM user_roles WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("leagueRepo.GetAllUserRoles: %w", err)
	}
	return roles, nil
}

func (r *leagueRepo) ListLeagueRoles(ctx context.Context, leagueID int64) ([]model.UserRole, error) {
	roles := make([]model.UserRole, 0)
	err := r.db(ctx).SelectContext(ctx, &roles,
		`SELECT user_id, role_id, league_id FROM user_roles WHERE league_id = $1 ORDER BY user_id`,
		leagueID,
	)
	if err != nil {
		return nil, fmt.Errorf("leagueRepo.ListLeagueRoles: %w", err)
	}
	return roles, nil
}
