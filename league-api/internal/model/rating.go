package model

type RatingHistory struct {
	HistoryID  int64   `db:"history_id" json:"historyId"`
	UserID     int64   `db:"user_id"    json:"userId"`
	MatchID    int64   `db:"match_id"   json:"matchId"`
	Delta      float64 `db:"delta"      json:"delta"`
	Rating     float64 `db:"rating"     json:"rating"`
	Deviation  float64 `db:"deviation"  json:"deviation"`
	Volatility float64 `db:"volatility" json:"volatility"`
}
