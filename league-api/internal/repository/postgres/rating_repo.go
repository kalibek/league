package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

type ratingRepo struct {
	pool *sqlx.DB
}

func NewRatingRepo(db *sqlx.DB) repository.RatingRepository {
	return &ratingRepo{pool: db}
}

func (r *ratingRepo) db(ctx context.Context) idb.DBTX {
	if tx := idb.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *ratingRepo) InsertHistory(ctx context.Context, rh *model.RatingHistory) error {
	const q = `
		INSERT INTO rating_history (user_id, match_id, delta, rating, deviation, volatility)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db(ctx).ExecContext(ctx, q,
		rh.UserID, rh.MatchID, rh.Delta, rh.Rating, rh.Deviation, rh.Volatility,
	)
	if err != nil {
		return fmt.Errorf("ratingRepo.InsertHistory: %w", err)
	}
	return nil
}

func (r *ratingRepo) GetByUser(ctx context.Context, userID int64) ([]model.RatingHistory, error) {
	history := make([]model.RatingHistory, 0)
	err := r.db(ctx).SelectContext(ctx, &history,
		`SELECT * FROM rating_history WHERE user_id = $1 ORDER BY history_id DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("ratingRepo.GetByUser: %w", err)
	}
	return history, nil
}

func (r *ratingRepo) GetByUserInEvent(ctx context.Context, userID, eventID int64) ([]model.RatingHistory, error) {
	history := make([]model.RatingHistory, 0)
	err := r.db(ctx).SelectContext(ctx, &history, `
		SELECT rh.*
		FROM rating_history rh
		JOIN matches m ON rh.match_id = m.match_id
		JOIN groups g ON m.group_id = g.group_id
		WHERE g.event_id = $1 AND rh.user_id = $2
		ORDER BY rh.history_id ASC`,
		eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("ratingRepo.GetByUserInEvent: %w", err)
	}
	return history, nil
}

func (r *ratingRepo) GetEventDeltaForUser(ctx context.Context, userID, eventID int64) (float64, error) {
	var delta float64
	err := r.db(ctx).GetContext(ctx, &delta, `
		SELECT COALESCE(SUM(rh.delta), 0)
		FROM rating_history rh
		JOIN matches m ON rh.match_id = m.match_id
		JOIN groups g ON m.group_id = g.group_id
		WHERE g.event_id = $1 AND rh.user_id = $2`,
		eventID, userID)
	if err != nil {
		return 0, fmt.Errorf("ratingRepo.GetEventDeltaForUser: %w", err)
	}
	return delta, nil
}

func (r *ratingRepo) DeleteAll(ctx context.Context) error {
	_, err := r.db(ctx).ExecContext(ctx, `DELETE FROM rating_history`)
	if err != nil {
		return fmt.Errorf("ratingRepo.DeleteAll: %w", err)
	}
	return nil
}

func (r *ratingRepo) DeleteByGroup(ctx context.Context, groupID int64) error {
	const q = `
		DELETE FROM rating_history
		WHERE match_id IN (SELECT match_id FROM matches WHERE group_id = $1)`
	_, err := r.db(ctx).ExecContext(ctx, q, groupID)
	if err != nil {
		return fmt.Errorf("ratingRepo.DeleteByGroup: %w", err)
	}
	return nil
}
