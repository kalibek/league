package service

import (
	"context"
	"testing"
	"time"

	"league-api/internal/model"
)

// --- Additional profile service tests for 0% functions ---

func TestListCountries(t *testing.T) {
	pr := &mockProfileRepo{
		countries: []model.Country{
			{CountryID: 1, Name: "Kazakhstan", Code: "KZ"},
			{CountryID: 2, Name: "Russia", Code: "RU"},
		},
	}
	svc := NewProfileService(pr, &mockUserRepoForProfile{user: &model.User{}})

	countries, err := svc.ListCountries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(countries) != 2 {
		t.Errorf("expected 2 countries, got %d", len(countries))
	}
}

func TestListCities(t *testing.T) {
	pr := &mockProfileRepo{
		cities: []model.City{
			{CityID: 1, Name: "Almaty", CountryID: 1},
			{CityID: 2, Name: "Nur-Sultan", CountryID: 1},
		},
	}
	svc := NewProfileService(pr, &mockUserRepoForProfile{user: &model.User{}})

	cities, err := svc.ListCities(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cities) != 2 {
		t.Errorf("expected 2 cities, got %d", len(cities))
	}
}

func TestListBlades(t *testing.T) {
	pr := &mockProfileRepo{
		blades: []model.Blade{
			{BladeID: 1, Name: "Timo Boll ALC"},
			{BladeID: 2, Name: "Zhang Jike"},
		},
	}
	svc := NewProfileService(pr, &mockUserRepoForProfile{user: &model.User{}})

	blades, err := svc.ListBlades(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blades) != 2 {
		t.Errorf("expected 2 blades, got %d", len(blades))
	}
}

func TestListRubbers(t *testing.T) {
	pr := &mockProfileRepo{
		rubbers: []model.Rubber{
			{RubberID: 1, Name: "Tenergy 05"},
			{RubberID: 2, Name: "MXP"},
		},
	}
	svc := NewProfileService(pr, &mockUserRepoForProfile{user: &model.User{}})

	rubbers, err := svc.ListRubbers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rubbers) != 2 {
		t.Errorf("expected 2 rubbers, got %d", len(rubbers))
	}
}

func TestAddCity_Success(t *testing.T) {
	pr := &mockProfileRepo{}
	svc := NewProfileService(pr, &mockUserRepoForProfile{user: &model.User{}})

	city, err := svc.AddCity(context.Background(), "Shymkent", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if city == nil || city.Name != "Shymkent" {
		t.Errorf("expected city Shymkent, got %v", city)
	}
}

func TestAddBlade_Success(t *testing.T) {
	pr := &mockProfileRepo{}
	svc := NewProfileService(pr, &mockUserRepoForProfile{user: &model.User{}})

	blade, err := svc.AddBlade(context.Background(), "Viscaria")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blade == nil || blade.Name != "Viscaria" {
		t.Errorf("expected blade Viscaria, got %v", blade)
	}
}

func TestAddRubber_Success(t *testing.T) {
	pr := &mockProfileRepo{}
	svc := NewProfileService(pr, &mockUserRepoForProfile{user: &model.User{}})

	rubber, err := svc.AddRubber(context.Background(), "Dignics 09C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rubber == nil || rubber.Name != "Dignics 09C" {
		t.Errorf("expected rubber Dignics 09C, got %v", rubber)
	}
}

// --- NewMatchService construction test ---

func TestNewMatchService_Constructor(t *testing.T) {
	mr := &matchSvcMockMatchRepo{matches: map[int64]*model.Match{}, groupMatches: map[int64][]model.Match{}}
	gr := &matchSvcMockGroupRepo{groups: map[int64]*model.Group{}, players: map[int64][]model.GroupPlayer{}}
	hub := newTestHub()

	svc := NewMatchService(mr, gr, hub)
	if svc == nil {
		t.Fatal("expected non-nil match service")
	}
}

// --- NewDraftService construction test ---

func TestNewDraftService_Constructor(t *testing.T) {
	lr := &draftMockLeagueRepo{leagues: map[int64]*model.League{}}
	er := &draftMockEventRepo{events: map[int64]*model.LeagueEvent{}}
	gr := &draftMockGroupRepo{groups: map[int64][]model.Group{}, groupByID: map[int64]*model.Group{}}
	mr := &draftMockMatchRepo{matches: map[int64][]model.Match{}}

	svc := NewDraftService(lr, er, gr, mr, &nopMatchService{}, &nopRatingService{}, &nopGroupService{}, nil)
	if svc == nil {
		t.Fatal("expected non-nil draft service")
	}
}

// --- GetEventDetail error path test ---

func TestGetEventDetail_EventNotFound(t *testing.T) {
	er := &evtMockEventRepo{events: map[int64]*model.LeagueEvent{}}
	svc := NewEventService(er, &evtMockGroupRepo{groups: map[int64][]model.Group{}}, &evtMockMatchRepo{}, &evtMockUserRepo{})

	_, err := svc.GetEventDetail(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for missing event")
	}
}

func TestCreateDraftEvent_CreateError(t *testing.T) {
	er := &evtMockEventRepo{
		events:    map[int64]*model.LeagueEvent{},
		byLeague:  map[int64][]model.LeagueEvent{1: {}},
		createErr: errDummy,
	}
	svc := NewEventService(er, &evtMockGroupRepo{}, &evtMockMatchRepo{}, &evtMockUserRepo{})

	_, err := svc.CreateDraftEvent(context.Background(), 1, "Test", time.Now(), time.Now().Add(time.Hour))
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

// errDummy is a reusable sentinel error for coverage tests.
var errDummy = errFromString("injected error")

func errFromString(s string) error {
	return &dummyErr{s}
}

type dummyErr struct{ msg string }

func (e *dummyErr) Error() string { return e.msg }
