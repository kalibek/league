package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	idb "league-api/internal/db"
	"league-api/internal/model"
	"league-api/internal/repository"
)

type profileRepo struct {
	pool *sqlx.DB
}

func NewProfileRepo(db *sqlx.DB) repository.ProfileRepository {
	return &profileRepo{pool: db}
}

func (r *profileRepo) db(ctx context.Context) idb.DBTX {
	if tx := idb.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *profileRepo) GetProfile(ctx context.Context, userID int64) (*model.PlayerProfileRow, error) {
	var p model.PlayerProfileRow
	err := r.db(ctx).GetContext(ctx, &p,
		`SELECT * FROM player_profiles WHERE user_id = $1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("profileRepo.GetProfile: %w", err)
	}
	return &p, nil
}

func (r *profileRepo) UpsertProfile(ctx context.Context, p *model.PlayerProfileRow) error {
	const q = `
		INSERT INTO player_profiles
			(user_id, country_id, city_id, birthdate, grip, gender, blade_id, fh_rubber_id, bh_rubber_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			country_id   = EXCLUDED.country_id,
			city_id      = EXCLUDED.city_id,
			birthdate    = EXCLUDED.birthdate,
			grip         = EXCLUDED.grip,
			gender       = EXCLUDED.gender,
			blade_id     = EXCLUDED.blade_id,
			fh_rubber_id = EXCLUDED.fh_rubber_id,
			bh_rubber_id = EXCLUDED.bh_rubber_id,
			last_updated = NOW()`
	_, err := r.db(ctx).ExecContext(ctx, q,
		p.UserID, p.CountryID, p.CityID, p.Birthdate,
		p.Grip, p.Gender, p.BladeID, p.FhRubberID, p.BhRubberID)
	if err != nil {
		return fmt.Errorf("profileRepo.UpsertProfile: %w", err)
	}
	return nil
}

func (r *profileRepo) ListCountries(ctx context.Context) ([]model.Country, error) {
	var out []model.Country
	if err := r.db(ctx).SelectContext(ctx, &out,
		`SELECT country_id, name, code FROM countries ORDER BY name`); err != nil {
		return nil, fmt.Errorf("profileRepo.ListCountries: %w", err)
	}
	return out, nil
}

func (r *profileRepo) GetCountry(ctx context.Context, id int) (*model.Country, error) {
	var c model.Country
	if err := r.db(ctx).GetContext(ctx, &c,
		`SELECT country_id, name, code FROM countries WHERE country_id = $1`, id); err != nil {
		return nil, fmt.Errorf("profileRepo.GetCountry: %w", err)
	}
	return &c, nil
}

func (r *profileRepo) ListCities(ctx context.Context, countryID int) ([]model.City, error) {
	var out []model.City
	if err := r.db(ctx).SelectContext(ctx, &out,
		`SELECT city_id, name, country_id FROM cities WHERE country_id = $1 ORDER BY name`,
		countryID); err != nil {
		return nil, fmt.Errorf("profileRepo.ListCities: %w", err)
	}
	return out, nil
}

func (r *profileRepo) GetCity(ctx context.Context, id int) (*model.City, error) {
	var c model.City
	if err := r.db(ctx).GetContext(ctx, &c,
		`SELECT city_id, name, country_id FROM cities WHERE city_id = $1`, id); err != nil {
		return nil, fmt.Errorf("profileRepo.GetCity: %w", err)
	}
	return &c, nil
}

func (r *profileRepo) AddCity(ctx context.Context, name string, countryID int) (*model.City, error) {
	var c model.City
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO cities (name, country_id) VALUES ($1, $2)
			 ON CONFLICT (name, country_id) DO UPDATE SET name = EXCLUDED.name
			 RETURNING city_id, name, country_id`,
			name, countryID).Scan(&c.CityID, &c.Name, &c.CountryID)
	} else {
		err = r.pool.QueryRowContext(ctx,
			`INSERT INTO cities (name, country_id) VALUES ($1, $2)
			 ON CONFLICT (name, country_id) DO UPDATE SET name = EXCLUDED.name
			 RETURNING city_id, name, country_id`,
			name, countryID).Scan(&c.CityID, &c.Name, &c.CountryID)
	}
	if err != nil {
		return nil, fmt.Errorf("profileRepo.AddCity: %w", err)
	}
	return &c, nil
}

func (r *profileRepo) ListBlades(ctx context.Context) ([]model.Blade, error) {
	var out []model.Blade
	if err := r.db(ctx).SelectContext(ctx, &out,
		`SELECT blade_id, name FROM blades ORDER BY name`); err != nil {
		return nil, fmt.Errorf("profileRepo.ListBlades: %w", err)
	}
	return out, nil
}

func (r *profileRepo) GetBlade(ctx context.Context, id int) (*model.Blade, error) {
	var b model.Blade
	if err := r.db(ctx).GetContext(ctx, &b,
		`SELECT blade_id, name FROM blades WHERE blade_id = $1`, id); err != nil {
		return nil, fmt.Errorf("profileRepo.GetBlade: %w", err)
	}
	return &b, nil
}

func (r *profileRepo) AddBlade(ctx context.Context, name string) (*model.Blade, error) {
	var b model.Blade
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO blades (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING blade_id, name`,
			name).Scan(&b.BladeID, &b.Name)
	} else {
		err = r.pool.QueryRowContext(ctx,
			`INSERT INTO blades (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING blade_id, name`,
			name).Scan(&b.BladeID, &b.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("profileRepo.AddBlade: %w", err)
	}
	return &b, nil
}

func (r *profileRepo) ListRubbers(ctx context.Context) ([]model.Rubber, error) {
	var out []model.Rubber
	if err := r.db(ctx).SelectContext(ctx, &out,
		`SELECT rubber_id, name FROM rubbers ORDER BY name`); err != nil {
		return nil, fmt.Errorf("profileRepo.ListRubbers: %w", err)
	}
	return out, nil
}

func (r *profileRepo) GetRubber(ctx context.Context, id int) (*model.Rubber, error) {
	var rb model.Rubber
	if err := r.db(ctx).GetContext(ctx, &rb,
		`SELECT rubber_id, name FROM rubbers WHERE rubber_id = $1`, id); err != nil {
		return nil, fmt.Errorf("profileRepo.GetRubber: %w", err)
	}
	return &rb, nil
}

func (r *profileRepo) AddRubber(ctx context.Context, name string) (*model.Rubber, error) {
	var rb model.Rubber
	var err error
	if tx := idb.ExtractTx(ctx); tx != nil {
		err = tx.QueryRowContext(ctx,
			`INSERT INTO rubbers (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING rubber_id, name`,
			name).Scan(&rb.RubberID, &rb.Name)
	} else {
		err = r.pool.QueryRowContext(ctx,
			`INSERT INTO rubbers (name) VALUES ($1)
			 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			 RETURNING rubber_id, name`,
			name).Scan(&rb.RubberID, &rb.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("profileRepo.AddRubber: %w", err)
	}
	return &rb, nil
}
