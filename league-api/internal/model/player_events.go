package model

import "time"

type PlayerMatchSummary struct {
	MatchID      int64    `json:"matchId"`
	OpponentID   *int64   `json:"opponentId"`
	OpponentName string   `json:"opponentName"`
	MyScore      *int16   `json:"myScore"`
	OppScore     *int16   `json:"oppScore"`
	Won          *bool    `json:"won"`
	Withdraw     bool     `json:"withdraw"`
	OppWithdraw  bool     `json:"oppWithdraw"`
	Status       MatchStatus `json:"status"`
	RatingDelta  *float64 `json:"ratingDelta,omitempty"`
}

type PlayerGroupSummary struct {
	GroupID   int64                `json:"groupId"`
	Division  string               `json:"division"`
	GroupNo   int                  `json:"groupNo"`
	Status    GroupStatus          `json:"status"`
	Place     int16                `json:"place"`
	Points    int16                `json:"points"`
	Advances  bool                 `json:"advances"`
	Recedes   bool                 `json:"recedes"`
	Matches   []PlayerMatchSummary `json:"matches"`
}

type PlayerEventSummary struct {
	EventID      int64                `json:"eventId"`
	LeagueID     int64                `json:"leagueId"`
	Title        string               `json:"title"`
	StartDate    time.Time            `json:"startDate"`
	EndDate      time.Time            `json:"endDate"`
	Status       EventStatus          `json:"status"`
	RatingDelta  float64              `json:"ratingDelta"`
	RatingBefore *float64             `json:"ratingBefore,omitempty"`
	RatingAfter  *float64             `json:"ratingAfter,omitempty"`
	Groups       []PlayerGroupSummary `json:"groups"`
}

type PlayerEventsPage struct {
	Events []PlayerEventSummary `json:"events"`
	Total  int                  `json:"total"`
	Offset int                  `json:"offset"`
	Limit  int                  `json:"limit"`
}
