package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"league-api/internal/model"
	"league-api/internal/repository"
)

type matchRepo struct {
	db *sqlx.DB
}

func NewMatchRepo(db *sqlx.DB) repository.MatchRepository {
	return &matchRepo{db}
}

func (r *matchRepo) GetByID(ctx context.Context, id int64) (*model.Match, error) {
	var m model.Match
	err := r.db.GetContext(ctx, &m, `SELECT * FROM matches WHERE match_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("matchRepo.GetByID: %w", err)
	}
	return &m, nil
}

func (r *matchRepo) ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error) {
	matches := make([]model.Match, 0)
	err := r.db.SelectContext(ctx, &matches,
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
	err := r.db.QueryRowContext(ctx, q,
		m.GroupID, m.GroupPlayer1ID, m.GroupPlayer2ID, m.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("matchRepo.Create: %w", err)
	}
	return id, nil
}

func (r *matchRepo) UpdateScore(ctx context.Context, id int64, score1, score2 int16) error {
	const q = `
		UPDATE matches SET score1 = $1, score2 = $2, last_updated = NOW()
		WHERE match_id = $3`
	_, err := r.db.ExecContext(ctx, q, score1, score2, id)
	if err != nil {
		return fmt.Errorf("matchRepo.UpdateScore: %w", err)
	}
	return nil
}

func (r *matchRepo) UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error {
	const q = `UPDATE matches SET status = $1, last_updated = NOW() WHERE match_id = $2`
	_, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("matchRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *matchRepo) BulkCreate(ctx context.Context, matches []model.Match) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("matchRepo.BulkCreate begin: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `
		INSERT INTO matches (group_id, group_player1_id, group_player2_id, status)
		VALUES ($1, $2, $3, $4)`
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
	_, err := r.db.ExecContext(ctx, q, groupID)
	if err != nil {
		return fmt.Errorf("matchRepo.ResetGroupMatches: %w", err)
	}
	return nil
}

// SetWithdraw marks one side of a match as withdrawn and sets status=DONE.
// position must be 1 or 2.
func (r *matchRepo) SetWithdraw(ctx context.Context, matchID int64, position int) error {
	var q string
	switch position {
	case 1:
		q = `UPDATE matches SET withdraw1 = TRUE, status = 'DONE', last_updated = NOW() WHERE match_id = $1`
	case 2:
		q = `UPDATE matches SET withdraw2 = TRUE, status = 'DONE', last_updated = NOW() WHERE match_id = $1`
	default:
		return fmt.Errorf("matchRepo.SetWithdraw: invalid position %d", position)
	}
	_, err := r.db.ExecContext(ctx, q, matchID)
	if err != nil {
		return fmt.Errorf("matchRepo.SetWithdraw: %w", err)
	}
	return nil
}
