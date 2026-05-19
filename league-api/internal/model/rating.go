package model

// UserRatingSnapshot is the Glicko2 state for a user at a point in time.
// Used during partial rating recalculation to restore pre-event ratings.
type UserRatingSnapshot struct {
	UserID     int64   `db:"user_id"`
	Rating     float64 `db:"rating"`
	Deviation  float64 `db:"deviation"`
	Volatility float64 `db:"volatility"`
}

type RatingHistory struct {
	HistoryID  int64   `db:"history_id" json:"historyId"`
	UserID     int64   `db:"user_id"    json:"userId"`
	MatchID    int64   `db:"match_id"   json:"matchId"`
	Delta      float64 `db:"delta"      json:"delta"`
	Rating     float64 `db:"rating"     json:"rating"`
	Deviation  float64 `db:"deviation"  json:"deviation"`
	Volatility float64 `db:"volatility" json:"volatility"`
}
