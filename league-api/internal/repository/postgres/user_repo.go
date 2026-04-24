package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"league-api/internal/model"
	"league-api/internal/repository"
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) repository.UserRepository {
	return &userRepo{db}
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE user_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	return &u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetByEmail: %w", err)
	}
	return &u, nil
}

func (r *userRepo) Create(ctx context.Context, u *model.User) (int64, error) {
	const q = `
		INSERT INTO users (first_name, last_name, email, current_rating, deviation, volatility)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING user_id`
	var id int64
	err := r.db.QueryRowContext(ctx, q,
		u.FirstName, u.LastName, u.Email,
		u.CurrentRating, u.Deviation, u.Volatility,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("userRepo.Create: %w", err)
	}
	return id, nil
}

func (r *userRepo) List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error) {
	// Only allow safe sort columns to prevent SQL injection.
	orderCol := "current_rating DESC"
	switch sortBy {
	case "name":
		orderCol = "last_name ASC, first_name ASC"
	case "rating":
		orderCol = "current_rating DESC"
	}
	q := fmt.Sprintf(`SELECT * FROM users ORDER BY %s LIMIT $1 OFFSET $2`, orderCol)
	users := make([]model.User, 0)
	if err := r.db.SelectContext(ctx, &users, q, limit, offset); err != nil {
		return nil, fmt.Errorf("userRepo.List: %w", err)
	}
	return users, nil
}

func (r *userRepo) Search(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error) {
	orderCol := "u.current_rating DESC"
	switch sortBy {
	case "name":
		orderCol = "u.last_name ASC, u.first_name ASC"
	case "rating":
		orderCol = "u.current_rating DESC"
	}

	pattern := "%" + q + "%"
	query := fmt.Sprintf(`
		SELECT DISTINCT u.*
		FROM users u
		LEFT JOIN player_profiles pp ON pp.user_id = u.user_id
		LEFT JOIN countries c  ON c.country_id  = pp.country_id
		LEFT JOIN cities    ci ON ci.city_id     = pp.city_id
		LEFT JOIN blades    b  ON b.blade_id     = pp.blade_id
		LEFT JOIN rubbers   fhr ON fhr.rubber_id = pp.fh_rubber_id
		LEFT JOIN rubbers   bhr ON bhr.rubber_id = pp.bh_rubber_id
		WHERE
			u.first_name ILIKE $1 OR
			u.last_name  ILIKE $1 OR
			u.email      ILIKE $1 OR
			c.name       ILIKE $1 OR
			ci.name      ILIKE $1 OR
			pp.grip      ILIKE $1 OR
			pp.gender    ILIKE $1 OR
			b.name       ILIKE $1 OR
			fhr.name     ILIKE $1 OR
			bhr.name     ILIKE $1
		ORDER BY %s
		LIMIT $2 OFFSET $3`, orderCol)

	users := make([]model.User, 0)
	if err := r.db.SelectContext(ctx, &users, query, pattern, limit, offset); err != nil {
		return nil, fmt.Errorf("userRepo.Search: %w", err)
	}
	return users, nil
}

func (r *userRepo) UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error {
	const q = `
		UPDATE users
		SET current_rating = $1, deviation = $2, volatility = $3, last_updated = NOW()
		WHERE user_id = $4`
	_, err := r.db.ExecContext(ctx, q, rating, deviation, volatility, userID)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateRating: %w", err)
	}
	return nil
}

func (r *userRepo) ResetAllRatings(ctx context.Context) error {
	const q = `UPDATE users SET current_rating = 1500, deviation = 350, volatility = 0.06, last_updated = NOW()`
	_, err := r.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("userRepo.ResetAllRatings: %w", err)
	}
	return nil
}

func (r *userRepo) SetPasswordHash(ctx context.Context, userID int64, hash string) error {
	const q = `UPDATE users SET password_hash = $1, last_updated = NOW() WHERE user_id = $2`
	_, err := r.db.ExecContext(ctx, q, hash, userID)
	if err != nil {
		return fmt.Errorf("userRepo.SetPasswordHash: %w", err)
	}
	return nil
}

func (r *userRepo) UpdateName(ctx context.Context, userID int64, firstName, lastName string) error {
	const q = `UPDATE users SET first_name = $1, last_name = $2, last_updated = NOW() WHERE user_id = $3`
	_, err := r.db.ExecContext(ctx, q, firstName, lastName, userID)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateName: %w", err)
	}
	return nil
}
