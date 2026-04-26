package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"league-api/internal/model"
)

// --- mocks ---

type mockProfileRepo struct {
	profile   *model.PlayerProfileRow
	countries []model.Country
	cities    []model.City
	blades    []model.Blade
	rubbers   []model.Rubber
	upsertErr error
}

func (m *mockProfileRepo) GetProfile(_ context.Context, _ int64) (*model.PlayerProfileRow, error) {
	return m.profile, nil
}
func (m *mockProfileRepo) UpsertProfile(_ context.Context, p *model.PlayerProfileRow) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.profile = p
	return nil
}
func (m *mockProfileRepo) ListCountries(_ context.Context) ([]model.Country, error) {
	return m.countries, nil
}
func (m *mockProfileRepo) ListCities(_ context.Context, _ int) ([]model.City, error) {
	return m.cities, nil
}
func (m *mockProfileRepo) AddCity(_ context.Context, name string, countryID int) (*model.City, error) {
	c := &model.City{CityID: 99, Name: name, CountryID: countryID}
	return c, nil
}
func (m *mockProfileRepo) ListBlades(_ context.Context) ([]model.Blade, error) {
	return m.blades, nil
}
func (m *mockProfileRepo) AddBlade(_ context.Context, name string) (*model.Blade, error) {
	return &model.Blade{BladeID: 99, Name: name}, nil
}
func (m *mockProfileRepo) ListRubbers(_ context.Context) ([]model.Rubber, error) {
	return m.rubbers, nil
}
func (m *mockProfileRepo) AddRubber(_ context.Context, name string) (*model.Rubber, error) {
	return &model.Rubber{RubberID: 99, Name: name}, nil
}
func (m *mockProfileRepo) GetCountry(_ context.Context, id int) (*model.Country, error) {
	for _, c := range m.countries {
		if c.CountryID == id {
			return &c, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockProfileRepo) GetCity(_ context.Context, id int) (*model.City, error) {
	for _, c := range m.cities {
		if c.CityID == id {
			return &c, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockProfileRepo) GetBlade(_ context.Context, id int) (*model.Blade, error) {
	for _, b := range m.blades {
		if b.BladeID == id {
			return &b, nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockProfileRepo) GetRubber(_ context.Context, id int) (*model.Rubber, error) {
	for _, r := range m.rubbers {
		if r.RubberID == id {
			return &r, nil
		}
	}
	return nil, errors.New("not found")
}

type mockUserRepoForProfile struct {
	user      *model.User
	nameErr   error
	updateCalls int
}

func (m *mockUserRepoForProfile) GetByID(_ context.Context, _ int64) (*model.User, error) {
	return m.user, nil
}
func (m *mockUserRepoForProfile) GetByEmail(_ context.Context, _ string) (*model.User, error) { return nil, nil }
func (m *mockUserRepoForProfile) Create(_ context.Context, _ *model.User) (int64, error)     { return 0, nil }
func (m *mockUserRepoForProfile) List(_ context.Context, _, _ int, _ string) ([]model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForProfile) UpdateRating(_ context.Context, _ int64, _, _, _ float64) error {
	return nil
}
func (m *mockUserRepoForProfile) SetPasswordHash(_ context.Context, _ int64, _ string) error {
	return nil
}
func (m *mockUserRepoForProfile) UpdateName(_ context.Context, _ int64, _, _ string) error {
	m.updateCalls++
	return m.nameErr
}
func (m *mockUserRepoForProfile) Search(_ context.Context, _ string, _, _ int, _ string) ([]model.User, error) {
	return nil, nil
}
func (m *mockUserRepoForProfile) ResetAllRatings(_ context.Context) error { return nil }

// --- helpers ---

func intPtr(i int) *int    { return &i }
func strPtr(s string) *string { return &s }

func fullProfile(userID int64) *model.PlayerProfileRow {
	bd := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	cid := 1
	cityID := 2
	bladeID := 3
	fhID := 4
	bhID := 5
	grip := "shakehand"
	gender := "male"
	return &model.PlayerProfileRow{
		UserID:     userID,
		CountryID:  &cid,
		CityID:     &cityID,
		Birthdate:  &bd,
		Grip:       &grip,
		Gender:     &gender,
		BladeID:    &bladeID,
		FhRubberID: &fhID,
		BhRubberID: &bhID,
	}
}

// --- tests ---

func TestGetProfile_NoProfileRow_IsNotComplete(t *testing.T) {
	ur := &mockUserRepoForProfile{user: &model.User{FirstName: "John", LastName: "Doe"}}
	pr := &mockProfileRepo{}
	svc := NewProfileService(pr, ur)

	detail, err := svc.GetProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.IsComplete {
		t.Error("expected IsComplete=false when no profile row")
	}
	if detail.FirstName != "John" {
		t.Errorf("expected firstName John, got %s", detail.FirstName)
	}
}

func TestGetProfile_FullProfile_IsComplete(t *testing.T) {
	row := fullProfile(1)
	ur := &mockUserRepoForProfile{user: &model.User{FirstName: "Ann", LastName: "Smith"}}
	pr := &mockProfileRepo{
		profile: row,
		countries: []model.Country{{CountryID: 1, Name: "Kazakhstan", Code: "KZ"}},
		cities:    []model.City{{CityID: 2, Name: "Almaty", CountryID: 1}},
		blades:    []model.Blade{{BladeID: 3, Name: "Timo Boll ALC"}},
		rubbers:   []model.Rubber{{RubberID: 4, Name: "Tenergy 05"}, {RubberID: 5, Name: "MXP"}},
	}
	svc := NewProfileService(pr, ur)

	detail, err := svc.GetProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !detail.IsComplete {
		t.Error("expected IsComplete=true for full profile")
	}
	if detail.Country == nil || detail.Country.Name != "Kazakhstan" {
		t.Error("expected country Kazakhstan")
	}
}

func TestUpsertProfile_UpdatesName(t *testing.T) {
	ur := &mockUserRepoForProfile{user: &model.User{FirstName: "Old", LastName: "Name"}}
	pr := &mockProfileRepo{}
	svc := NewProfileService(pr, ur)

	req := UpsertProfileRequest{FirstName: "New", LastName: "Name"}
	_, err := svc.UpsertProfile(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ur.updateCalls != 1 {
		t.Errorf("expected UpdateName called once, got %d", ur.updateCalls)
	}
}

func TestUpsertProfile_InvalidBirthdate(t *testing.T) {
	ur := &mockUserRepoForProfile{user: &model.User{}}
	pr := &mockProfileRepo{}
	svc := NewProfileService(pr, ur)

	bad := "not-a-date"
	req := UpsertProfileRequest{Birthdate: &bad}
	_, err := svc.UpsertProfile(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for invalid birthdate")
	}
}

func TestUpsertProfile_RepoError(t *testing.T) {
	ur := &mockUserRepoForProfile{user: &model.User{}}
	pr := &mockProfileRepo{upsertErr: errors.New("db error")}
	svc := NewProfileService(pr, ur)

	req := UpsertProfileRequest{}
	_, err := svc.UpsertProfile(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestAddCity_EmptyName(t *testing.T) {
	svc := NewProfileService(&mockProfileRepo{}, &mockUserRepoForProfile{})
	_, err := svc.AddCity(context.Background(), "", 1)
	if err == nil {
		t.Fatal("expected error for empty city name")
	}
}

func TestAddBlade_EmptyName(t *testing.T) {
	svc := NewProfileService(&mockProfileRepo{}, &mockUserRepoForProfile{})
	_, err := svc.AddBlade(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty blade name")
	}
}

func TestAddRubber_EmptyName(t *testing.T) {
	svc := NewProfileService(&mockProfileRepo{}, &mockUserRepoForProfile{})
	_, err := svc.AddRubber(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty rubber name")
	}
}

func TestIsProfileComplete_MissingFields(t *testing.T) {
	cases := []struct {
		name   string
		detail model.PlayerProfileDetail
		want   bool
	}{
		{"empty", model.PlayerProfileDetail{}, false},
		{"no last name", model.PlayerProfileDetail{FirstName: "X"}, false},
		{"names only", model.PlayerProfileDetail{FirstName: "X", LastName: "Y"}, false},
	}
	for _, tc := range cases {
		got := isProfileComplete(&tc.detail)
		if got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
