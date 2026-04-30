package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

type groupRepo struct {
	pool *sqlx.DB
}

func NewGroupRepo(db *sqlx.DB) repository.GroupRepository {
	return &groupRepo{pool: db}
}

func (r *groupRepo) db(ctx context.Context) idb.DBTX {
	if tx := idb.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *groupRepo) GetByID(ctx context.Context, id int64) (*model.Group, error) {
	var g model.Group
	err := r.db(ctx).GetContext(ctx, &g, `SELECT * FROM groups WHERE group_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("groupRepo.GetByID: %w", err)
	}
	return &g, nil
}

func (r *groupRepo) ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error) {
	groups := make([]model.Group, 0)
	err := r.db(ctx).SelectContext(ctx, &groups,
		`SELECT * FROM groups WHERE event_id = $1 ORDER BY division, group_no`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("groupRepo.ListByEvent: %w", err)
	}
	return groups, nil
}

func (r *groupRepo) Create(ctx context.Context, g *model.Group) (int64, error) {
	const q = `
		INSERT INTO groups (event_id, status, division, group_no, scheduled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING group_id`
	var id int64
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx, q,
			g.EventID, g.Status, g.Division, g.GroupNo, g.Scheduled,
		).Scan(&id)
	} else {
		err = r.pool.QueryRowContext(ctx, q,
			g.EventID, g.Status, g.Division, g.GroupNo, g.Scheduled,
		).Scan(&id)
	}
	if err != nil {
		return 0, fmt.Errorf("groupRepo.Create: %w", err)
	}
	return id, nil
}

func (r *groupRepo) UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error {
	const q = `UPDATE groups SET status = $1, last_updated = NOW() WHERE group_id = $2`
	_, err := r.db(ctx).ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("groupRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *groupRepo) GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error) {
	players := make([]model.GroupPlayer, 0)
	err := r.db(ctx).SelectContext(ctx, &players,
		`SELECT * FROM group_players WHERE group_id = $1 ORDER BY seed`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("groupRepo.GetPlayers: %w", err)
	}
	return players, nil
}

func (r *groupRepo) GetPlayersByMovement(ctx context.Context, groupID int64, movement int) ([]model.GroupPlayer, error) {
	query := ""
	switch movement {
	case model.MoveUp:
		query = `SELECT * FROM group_players WHERE group_id = $1 AND advances = true ORDER BY place`
		break
	case model.MoveDown:
		query = `SELECT * FROM group_players WHERE group_id = $1 AND recedes = true ORDER BY place DESC`
		break
	case model.MoveStay:
		query = `SELECT * FROM group_players WHERE group_id = $1 AND advances = false and recedes = false ORDER BY place`
	}

	players := make([]model.GroupPlayer, 0)
	err := r.db(ctx).SelectContext(ctx, &players,
		query,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("groupRepo.GetPlayers: %w", err)
	}
	return players, nil
}

func (r *groupRepo) AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error) {
	const q = `
		INSERT INTO group_players (group_id, user_id, seed, is_non_calculated)
		VALUES ($1, $2, $3, $4)
		RETURNING group_player_id`
	var id int64
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx, q,
			gp.GroupID, gp.UserID, gp.Seed, gp.IsNonCalculated,
		).Scan(&id)
	} else {
		err = r.pool.QueryRowContext(ctx, q,
			gp.GroupID, gp.UserID, gp.Seed, gp.IsNonCalculated,
		).Scan(&id)
	}
	if err != nil {
		return 0, fmt.Errorf("groupRepo.AddPlayer: %w", err)
	}
	return id, nil
}

func (r *groupRepo) RemovePlayer(ctx context.Context, groupPlayerID int64) error {
	_, err := r.db(ctx).ExecContext(ctx, `DELETE FROM group_players WHERE group_player_id = $1`, groupPlayerID)
	if err != nil {
		return fmt.Errorf("groupRepo.RemovePlayer: %w", err)
	}
	return nil
}

func (r *groupRepo) ResetGroupPlayers(ctx context.Context, groupID int64) error {
	const q = `
		UPDATE group_players
		SET place = 0, points = 0, tiebreak_points = 0,
		    advances = FALSE, recedes = FALSE, last_updated = NOW()
		WHERE group_id = $1`
	_, err := r.db(ctx).ExecContext(ctx, q, groupID)
	if err != nil {
		return fmt.Errorf("groupRepo.ResetGroupPlayers: %w", err)
	}
	return nil
}

func (r *groupRepo) ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error) {
	players := make([]model.GroupPlayer, 0)
	err := r.db(ctx).SelectContext(ctx, &players, `
		SELECT gp.*
		FROM group_players gp
		JOIN groups g ON g.group_id = gp.group_id
		WHERE gp.user_id = $1 AND g.event_id = $2`,
		userID, eventID)
	if err != nil {
		return nil, fmt.Errorf("groupRepo.ListPlayerGroupsInEvent: %w", err)
	}
	return players, nil
}

func (r *groupRepo) UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error {
	const q = `
		UPDATE group_players
		SET place = $1, points = $2, tiebreak_points = $3, advances = $4, recedes = $5, last_updated = NOW()
		WHERE group_player_id = $6`
	_, err := r.db(ctx).ExecContext(ctx, q,
		gp.Place, gp.Points, gp.TiebreakPoints, gp.Advances, gp.Recedes,
		gp.GroupPlayerID,
	)
	if err != nil {
		return fmt.Errorf("groupRepo.UpdatePlayer: %w", err)
	}
	return nil
}

func (r *groupRepo) SetPlayerStatus(ctx context.Context, groupPlayerID int64, status model.PlayerStatus) error {
	const q = `
		UPDATE group_players
		SET player_status = $1, last_updated = NOW()
		WHERE group_player_id = $2`
	_, err := r.db(ctx).ExecContext(ctx, q, string(status), groupPlayerID)
	if err != nil {
		return fmt.Errorf("groupRepo.SetPlayerStatus: %w", err)
	}
	return nil
}

func (r *groupRepo) ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error) {
	// Only allow safe sort columns to prevent SQL injection.
	q, args, err := sqlx.In(`SELECT * FROM users WHERE user_id in (?) ORDER BY current_rating DESC`, ids)
	if err != nil {
		return nil, fmt.Errorf("groupRepo.ListUsersByIdsByRatingDesc in query: %w", err)
	}
	q = r.pool.Rebind(q)

	users := make([]model.User, 0)
	if err := r.db(ctx).SelectContext(ctx, &users, q, args...); err != nil {
		return nil, fmt.Errorf("groupRepo.ListUsersByIdsByRatingDesc select query: %w", err)
	}
	return users, nil
}
