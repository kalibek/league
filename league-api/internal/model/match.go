package model

import "time"

type MatchStatus string

const (
	MatchDraft      MatchStatus = "DRAFT"
	MatchInProgress MatchStatus = "IN_PROGRESS"
	MatchDone       MatchStatus = "DONE"
)

type Match struct {
	MatchID        int64       `db:"match_id"         json:"matchId"`
	GroupID        int64       `db:"group_id"         json:"groupId"`
	GroupPlayer1ID *int64      `db:"group_player1_id" json:"groupPlayer1Id"`
	GroupPlayer2ID *int64      `db:"group_player2_id" json:"groupPlayer2Id"`
	Score1         *int16      `db:"score1"           json:"score1"`
	Score2         *int16      `db:"score2"           json:"score2"`
	Withdraw1      bool        `db:"withdraw1"        json:"withdraw1"`
	Withdraw2      bool        `db:"withdraw2"        json:"withdraw2"`
	Status         MatchStatus `db:"status"           json:"status"`
	Created        time.Time   `db:"created"          json:"created"`
	LastUpdated    time.Time   `db:"last_updated"     json:"lastUpdated"`
}
