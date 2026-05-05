package repository

import (
	"context"
	"time"

	"league-api/internal/model"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, u *model.User) (int64, error)
	List(ctx context.Context, limit, offset int, sortBy string) ([]model.User, error)
	Search(ctx context.Context, q string, limit, offset int, sortBy string) ([]model.User, error)
	UpdateRating(ctx context.Context, userID int64, rating, deviation, volatility float64) error
	ResetAllRatings(ctx context.Context) error
	SetPasswordHash(ctx context.Context, userID int64, hash string) error
	UpdateName(ctx context.Context, userID int64, firstName, lastName string) error
}

type OAuthAccountRepository interface {
	GetByProviderSub(ctx context.Context, provider, sub string) (*model.OAuthAccount, error)
	Create(ctx context.Context, a *model.OAuthAccount) error
	ListByUser(ctx context.Context, userID int64) ([]model.OAuthAccount, error)
}

type LeagueRepository interface {
	GetByID(ctx context.Context, id int64) (*model.League, error)
	List(ctx context.Context) ([]model.League, error)
	ListWithStats(ctx context.Context) ([]model.LeagueWithStats, error)
	ListMaintainers(ctx context.Context, leagueID int64, limit int) ([]model.LeagueMaintainer, error)
	Create(ctx context.Context, l *model.League) (int64, error)
	UpdateConfig(ctx context.Context, id int64, config model.LeagueConfig) error
	AssignRole(ctx context.Context, ur model.UserRole) error
	RemoveRole(ctx context.Context, userID, leagueID int64, roleID int) error
	GetUserRoles(ctx context.Context, userID, leagueID int64) ([]model.UserRole, error)
	GetAllUserRoles(ctx context.Context, userID int64) ([]model.UserRole, error)
	ListLeagueRoles(ctx context.Context, leagueID int64) ([]model.UserRole, error)
}

type EventRepository interface {
	GetByID(ctx context.Context, id int64) (*model.LeagueEvent, error)
	ListByLeague(ctx context.Context, leagueID int64) ([]model.LeagueEvent, error)
	ListDone(ctx context.Context) ([]model.LeagueEvent, error)
	Create(ctx context.Context, e *model.LeagueEvent) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status model.EventStatus) error
	ListEventsForPlayer(ctx context.Context, userID int64, limit, offset int) ([]model.LeagueEvent, int, error)
	UpdateDetails(ctx context.Context, id int64, title string, startDate, endDate time.Time) error
}

type GroupRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Group, error)
	ListByEvent(ctx context.Context, eventID int64) ([]model.Group, error)
	Create(ctx context.Context, g *model.Group) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status model.GroupStatus) error
	GetPlayers(ctx context.Context, groupID int64) ([]model.GroupPlayer, error)
	GetPlayersByMovement(ctx context.Context, groupID int64, moves int) ([]model.GroupPlayer, error)
	AddPlayer(ctx context.Context, gp *model.GroupPlayer) (int64, error)
	UpdatePlayer(ctx context.Context, gp *model.GroupPlayer) error
	SetPlayerStatus(ctx context.Context, groupPlayerID int64, status model.PlayerStatus) error
	RemovePlayer(ctx context.Context, groupPlayerID int64) error
	ResetGroupPlayers(ctx context.Context, groupID int64) error
	ListPlayerGroupsInEvent(ctx context.Context, userID, eventID int64) ([]model.GroupPlayer, error)
	ListUsersByIdsByRatingDesc(ctx context.Context, ids []int64) ([]model.User, error)
}

type MatchRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Match, error)
	ListByGroup(ctx context.Context, groupID int64) ([]model.Match, error)
	Create(ctx context.Context, m *model.Match) (int64, error)
	UpdateScore(ctx context.Context, id int64, score1, score2 int16, withdraw1, withdraw2 bool) error
	UpdateStatus(ctx context.Context, id int64, status model.MatchStatus) error
	BulkCreate(ctx context.Context, matches []model.Match) error
	ResetGroupMatches(ctx context.Context, groupID int64) error
	SetTableNumber(ctx context.Context, matchID int64, tableNumber int) error
	ResetScore(ctx context.Context, matchID int64) error
	ListInProgressByEvent(ctx context.Context, eventID int64) ([]int, error)
}

type ProfileRepository interface {
	GetProfile(ctx context.Context, userID int64) (*model.PlayerProfileRow, error)
	UpsertProfile(ctx context.Context, p *model.PlayerProfileRow) error
	ListCountries(ctx context.Context) ([]model.Country, error)
	ListCities(ctx context.Context, countryID int) ([]model.City, error)
	AddCity(ctx context.Context, name string, countryID int) (*model.City, error)
	ListBlades(ctx context.Context) ([]model.Blade, error)
	AddBlade(ctx context.Context, name string) (*model.Blade, error)
	ListRubbers(ctx context.Context) ([]model.Rubber, error)
	AddRubber(ctx context.Context, name string) (*model.Rubber, error)
	GetCountry(ctx context.Context, id int) (*model.Country, error)
	GetCity(ctx context.Context, id int) (*model.City, error)
	GetBlade(ctx context.Context, id int) (*model.Blade, error)
	GetRubber(ctx context.Context, id int) (*model.Rubber, error)
}

type RatingRepository interface {
	InsertHistory(ctx context.Context, rh *model.RatingHistory) error
	GetByUser(ctx context.Context, userID int64) ([]model.RatingHistory, error)
	GetByUserInEvent(ctx context.Context, userID, eventID int64) ([]model.RatingHistory, error)
	DeleteByGroup(ctx context.Context, groupID int64) error
	DeleteAll(ctx context.Context) error
	GetEventDeltaForUser(ctx context.Context, userID, eventID int64) (float64, error)
}
