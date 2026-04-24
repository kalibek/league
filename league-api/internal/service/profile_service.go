package service

import (
	"context"
	"fmt"
	"time"

	"league-api/internal/model"
	"league-api/internal/repository"
)

type UpsertProfileRequest struct {
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	CountryID   *int    `json:"countryId"`
	CityID      *int    `json:"cityId"`
	Birthdate   *string `json:"birthdate"` // "YYYY-MM-DD"
	Grip        *string `json:"grip"`
	Gender      *string `json:"gender"`
	BladeID     *int    `json:"bladeId"`
	FhRubberID  *int    `json:"fhRubberId"`
	BhRubberID  *int    `json:"bhRubberId"`
}

type ProfileService interface {
	GetProfile(ctx context.Context, userID int64) (*model.PlayerProfileDetail, error)
	UpsertProfile(ctx context.Context, userID int64, req UpsertProfileRequest) (*model.PlayerProfileDetail, error)
	ListCountries(ctx context.Context) ([]model.Country, error)
	ListCities(ctx context.Context, countryID int) ([]model.City, error)
	AddCity(ctx context.Context, name string, countryID int) (*model.City, error)
	ListBlades(ctx context.Context) ([]model.Blade, error)
	AddBlade(ctx context.Context, name string) (*model.Blade, error)
	ListRubbers(ctx context.Context) ([]model.Rubber, error)
	AddRubber(ctx context.Context, name string) (*model.Rubber, error)
}

type profileService struct {
	profileRepo repository.ProfileRepository
	userRepo    repository.UserRepository
}

func NewProfileService(profileRepo repository.ProfileRepository, userRepo repository.UserRepository) ProfileService {
	return &profileService{profileRepo: profileRepo, userRepo: userRepo}
}

func (s *profileService) GetProfile(ctx context.Context, userID int64) (*model.PlayerProfileDetail, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profileService.GetProfile user: %w", err)
	}

	row, err := s.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profileService.GetProfile row: %w", err)
	}

	detail := &model.PlayerProfileDetail{
		UserID:    userID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	if row != nil {
		detail.Birthdate = row.Birthdate
		detail.Grip = row.Grip
		detail.Gender = row.Gender

		if row.CountryID != nil {
			c, _ := s.profileRepo.GetCountry(ctx, *row.CountryID)
			detail.Country = c
		}
		if row.CityID != nil {
			ci, _ := s.profileRepo.GetCity(ctx, *row.CityID)
			detail.City = ci
		}
		if row.BladeID != nil {
			b, _ := s.profileRepo.GetBlade(ctx, *row.BladeID)
			detail.Blade = b
		}
		if row.FhRubberID != nil {
			rb, _ := s.profileRepo.GetRubber(ctx, *row.FhRubberID)
			detail.FhRubber = rb
		}
		if row.BhRubberID != nil {
			rb, _ := s.profileRepo.GetRubber(ctx, *row.BhRubberID)
			detail.BhRubber = rb
		}
	}

	detail.IsComplete = isProfileComplete(detail)
	return detail, nil
}

func (s *profileService) UpsertProfile(ctx context.Context, userID int64, req UpsertProfileRequest) (*model.PlayerProfileDetail, error) {
	if req.FirstName != "" || req.LastName != "" {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("profileService.UpsertProfile get user: %w", err)
		}
		fn, ln := user.FirstName, user.LastName
		if req.FirstName != "" {
			fn = req.FirstName
		}
		if req.LastName != "" {
			ln = req.LastName
		}
		if err := s.userRepo.UpdateName(ctx, userID, fn, ln); err != nil {
			return nil, fmt.Errorf("profileService.UpsertProfile update name: %w", err)
		}
	}

	row := &model.PlayerProfileRow{
		UserID:     userID,
		CountryID:  req.CountryID,
		CityID:     req.CityID,
		Grip:       req.Grip,
		Gender:     req.Gender,
		BladeID:    req.BladeID,
		FhRubberID: req.FhRubberID,
		BhRubberID: req.BhRubberID,
	}
	if req.Birthdate != nil && *req.Birthdate != "" {
		t, err := time.Parse("2006-01-02", *req.Birthdate)
		if err != nil {
			return nil, fmt.Errorf("invalid birthdate format (expected YYYY-MM-DD): %w", err)
		}
		row.Birthdate = &t
	}

	if err := s.profileRepo.UpsertProfile(ctx, row); err != nil {
		return nil, fmt.Errorf("profileService.UpsertProfile upsert: %w", err)
	}

	return s.GetProfile(ctx, userID)
}

func (s *profileService) ListCountries(ctx context.Context) ([]model.Country, error) {
	return s.profileRepo.ListCountries(ctx)
}

func (s *profileService) ListCities(ctx context.Context, countryID int) ([]model.City, error) {
	return s.profileRepo.ListCities(ctx, countryID)
}

func (s *profileService) AddCity(ctx context.Context, name string, countryID int) (*model.City, error) {
	if name == "" {
		return nil, fmt.Errorf("city name required")
	}
	return s.profileRepo.AddCity(ctx, name, countryID)
}

func (s *profileService) ListBlades(ctx context.Context) ([]model.Blade, error) {
	return s.profileRepo.ListBlades(ctx)
}

func (s *profileService) AddBlade(ctx context.Context, name string) (*model.Blade, error) {
	if name == "" {
		return nil, fmt.Errorf("blade name required")
	}
	return s.profileRepo.AddBlade(ctx, name)
}

func (s *profileService) ListRubbers(ctx context.Context) ([]model.Rubber, error) {
	return s.profileRepo.ListRubbers(ctx)
}

func (s *profileService) AddRubber(ctx context.Context, name string) (*model.Rubber, error) {
	if name == "" {
		return nil, fmt.Errorf("rubber name required")
	}
	return s.profileRepo.AddRubber(ctx, name)
}

func isProfileComplete(d *model.PlayerProfileDetail) bool {
	return d.FirstName != "" &&
		d.LastName != "" &&
		d.Country != nil &&
		d.City != nil &&
		d.Birthdate != nil &&
		d.Grip != nil &&
		d.Gender != nil &&
		d.Blade != nil &&
		d.FhRubber != nil &&
		d.BhRubber != nil
}
