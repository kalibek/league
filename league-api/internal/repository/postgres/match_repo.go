package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

type matchRepo struct {
	pool *sqlx.DB
}

func NewMatchRepo(db *sqlx.DB) repository.MatchRepository {
	return &matchRepo{pool: db}
}

func (r *matchRepo) db(ctx context.Context) idb.DBTX {
	if tx := idb.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *matchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) {
	var m model.Match
	err := r.db(ctx).GetContext(ctx, &m, `SELECT * FROM matches WHERE match_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("matchRepo.GetByID: %w", err)
	}
	return &m, nil
}

func (r *matchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	matches := make([]model.Match, 0)
	err := r.db(ctx).SelectContext(ctx, &matches,
		`SELECT * FROM matches WHERE group_id = $1 ORDER BY match_id`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("matchRepo.ListByGroup: %w", err)
	}
	return matches, nil
}

func (r *matchRepo) Create(ctx context.Context, m *model.Match) (int64, error) {
	const q = `
		INSERT INTO matches (group_id, group_player1_id, group_player2_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING match_id`
	var id int64
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx, q,
			m.GroupID, m.GroupPlayer1ID, m.GroupPlayer2ID, m.Status,
		).Scan(&id)
	} else {
		err = r.pool.QueryRowContext(ctx, q,
			m.GroupID, m.GroupPlayer1ID, m.GroupPlayer2ID, m.Status,
		).Scan(&id)
	}
	if err != nil {
		return 0, fmt.Errorf("matchRepo.Create: %w", err)
	}
	return id, nil
}

func (r *matchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error {
	const q = `
		UPDATE matches SET score1 = $1, score2 = $2,
		    withdraw1 = $3, withdraw2 = $4,
			last_updated = NOW()
		WHERE match_id = $5`
	_, err := r.db(ctx).ExecContext(ctx, q, score1, score2, withdraw1, withdraw2, id)
	if err != nil {
		return fmt.Errorf("matchRepo.UpdateScore: %w", err)
	}
	return nil
}

func (r *matchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	const q = `UPDATE matches SET status = $1, last_updated = NOW() WHERE match_id = $2`
	_, err := r.db(ctx).ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("matchRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *matchRepo) BulkCreate(ctx context.Context, matches []model.Match) error {
	const q = `
		INSERT INTO matches (group_id, group_player1_id, group_player2_id, status)
		VALUES ($1, $2, $3, $4)`

	// If a transaction is already in context, use it directly without starting a new one.
	if tx := idb.ExtractTx(ctx); tx != nil {
		for _, m := range matches {
			if _, err := tx.ExecContext(ctx, q, m.GroupID, m.GroupPlayer1ID, m.GroupPlayer2ID, m.Status); err != nil {
				return fmt.Errorf("matchRepo.BulkCreate insert: %w", err)
			}
		}
		return nil
	}

	tx, err := r.pool.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("matchRepo.BulkCreate begin: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for _, m := range matches {
		if _, err := tx.ExecContext(ctx, q, m.GroupID, m.GroupPlayer1ID, m.GroupPlayer2ID, m.Status); err != nil {
			return fmt.Errorf("matchRepo.BulkCreate insert: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("matchRepo.BulkCreate commit: %w", err)
	}
	return nil
}

func (r *matchRepo) ResetGroupMatches(ctx context.Context, groupID int64) error {
	const q = `
		UPDATE matches
		SET score1 = NULL, score2 = NULL,
		    withdraw1 = FALSE, withdraw2 = FALSE,
		    status = 'DRAFT', last_updated = NOW()
		WHERE group_id = $1`
	_, err := r.db(ctx).ExecContext(ctx, q, groupID)
	if err != nil {
		return fmt.Errorf("matchRepo.ResetGroupMatches: %w", err)
	}
	return nil
}

func (r *matchRepo) SetTableNumber(ctx context.Context, matchID int64, tableNumber int) error {
	const q = `
		UPDATE matches SET table_number = $1, status = 'IN_PROGRESS', last_updated = NOW()
		WHERE match_id = $2`
	_, err := r.db(ctx).ExecContext(ctx, q, tableNumber, matchID)
	if err != nil {
		return fmt.Errorf("matchRepo.SetTableNumber: %w", err)
	}
	return nil
}

func (r *matchRepo) ResetScore(ctx context.Context, matchID int64) error {
	const q = `
		UPDATE matches
		SET score1 = NULL, score2 = NULL, withdraw1 = false, withdraw2 = false,
		    table_number = NULL, status = 'DRAFT', last_updated = NOW()
		WHERE match_id = $1`
	_, err := r.db(ctx).ExecContext(ctx, q, matchID)
	if err != nil {
		return fmt.Errorf("matchRepo.ResetScore: %w", err)
	}
	return nil
}

func (r *matchRepo) ListInProgressByEvent(ctx context.Context, eventID int64) ([]int, error) {
	tables := make([]int, 0)
	const q = `
		SELECT m.table_number FROM matches m
		JOIN groups g ON g.group_id = m.group_id
		WHERE g.event_id = $1 AND m.status = 'IN_PROGRESS'`
	err := r.db(ctx).SelectContext(ctx, &tables, q, eventID)
	if err != nil {
		return nil, fmt.Errorf("matchRepo.ListInProgressByEvent: %w", err)
	}
	return tables, nil
}
