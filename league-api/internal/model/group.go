package model

import "time"

type GroupStatus string

const (
	GroupDraft      GroupStatus = "DRAFT"
	GroupInProgress GroupStatus = "IN_PROGRESS"
	GroupDone       GroupStatus = "DONE"
)

const (
	MoveUp = iota
	MoveDown
	MoveStay
)

type Group struct {
	GroupID     int64       `db:"group_id"     json:"groupId"`
	EventID     int64       `db:"event_id"     json:"eventId"`
	Status      GroupStatus `db:"status"       json:"status"`
	Division    string      `db:"division"     json:"division"`
	GroupNo     int         `db:"group_no"     json:"groupNo"`
	Scheduled   time.Time   `db:"scheduled"    json:"scheduled"`
	Created     time.Time   `db:"created"      json:"created"`
	LastUpdated time.Time   `db:"last_updated" json:"lastUpdated"`
}

type GroupDetail struct {
	Group
	Players []GroupPlayer `json:"players"`
	Matches []Match       `json:"matches"`
}

type EventDetail struct {
	LeagueEvent
	Groups []GroupDetail `json:"groups"`
}

type GroupPlayer struct {
	GroupPlayerID   int64     `db:"group_player_id"   json:"groupPlayerId"`
	GroupID         int64     `db:"group_id"          json:"groupId"`
	UserID          int64     `db:"user_id"           json:"userId"`
	Seed            int16     `db:"seed"              json:"seed"`
	Place           int16     `db:"place"             json:"place"`
	Points          int16     `db:"points"            json:"points"`
	TiebreakPoints  int16     `db:"tiebreak_points"   json:"tiebreakPoints"`
	Advances        bool      `db:"advances"          json:"advances"`
	Recedes         bool      `db:"recedes"           json:"recedes"`
	IsNonCalculated bool      `db:"is_non_calculated" json:"isNonCalculated"`
	User            *User     `db:"-"                 json:"user,omitempty"`
	Created         time.Time `db:"created"           json:"created"`
	LastUpdated     time.Time `db:"last_updated"      json:"lastUpdated"`
}
