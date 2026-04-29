package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

type eventRepo struct {
	pool *sqlx.DB
}

func NewEventRepo(db *sqlx.DB) repository.EventRepository {
	return &eventRepo{pool: db}
}

func (r *eventRepo) db(ctx context.Context) idb.DBTX {
	if tx := idb.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *eventRepo) GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error) {
	var e model.LeagueEvent
	err := r.db(ctx).GetContext(ctx, &e, `SELECT * FROM league_events WHERE event_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("eventRepo.GetByID: %w", err)
	}
	return &e, nil
}

func (r *eventRepo) ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error) {
	events := make([]model.LeagueEvent, 0)
	err := r.db(ctx).SelectContext(ctx, &events,
		`SELECT * FROM league_events WHERE league_id = $1 ORDER BY start_date DESC`,
		leagueID,
	)
	if err != nil {
		return nil, fmt.Errorf("eventRepo.ListByLeague: %w", err)
	}
	return events, nil
}

func (r *eventRepo) Create(ctx context.Context, e *model.LeagueEvent) (int64, error) {
	const q = `
		INSERT INTO league_events (league_id, status, title, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING event_id`
	var id int64
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx, q,
			e.LeagueID, e.Status, e.Title, e.StartDate, e.EndDate,
		).Scan(&id)
	} else {
		err = r.pool.QueryRowContext(ctx, q,
			e.LeagueID, e.Status, e.Title, e.StartDate, e.EndDate,
		).Scan(&id)
	}
	if err != nil {
		return 0, fmt.Errorf("eventRepo.Create: %w", err)
	}
	return id, nil
}

func (r *eventRepo) UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error {
	const q = `UPDATE league_events SET status = $1, last_updated = NOW() WHERE event_id = $2`
	_, err := r.db(ctx).ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("eventRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *eventRepo) ListDone(ctx context.Context) ([]model.LeagueEvent, error) {
	events := make([]model.LeagueEvent, 0)
	err := r.db(ctx).SelectContext(ctx, &events,
		`SELECT * FROM league_events WHERE status = 'DONE' ORDER BY start_date ASC, event_id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("eventRepo.ListDone: %w", err)
	}
	return events, nil
}

func (r *eventRepo) ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error) {
	var total int
	err := r.db(ctx).GetContext(ctx, &total, `
		SELECT COUNT(DISTINCT le.event_id)
		FROM league_events le
		JOIN groups g ON g.event_id = le.event_id
		JOIN group_players gp ON gp.group_id = g.group_id
		WHERE gp.user_id = $1`, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("eventRepo.ListEventsForPlayer count: %w", err)
	}

	events := make([]model.LeagueEvent, 0)
	err = r.db(ctx).SelectContext(ctx, &events, `
		SELECT DISTINCT le.*
		FROM league_events le
		JOIN groups g ON g.event_id = le.event_id
		JOIN group_players gp ON gp.group_id = g.group_id
		WHERE gp.user_id = $1
		ORDER BY le.end_date DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("eventRepo.ListEventsForPlayer list: %w", err)
	}
	return events, total, nil
}
