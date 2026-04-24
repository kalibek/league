package model

import "time"

type EventStatus string

const (
	EventDraft      EventStatus = "DRAFT"
	EventInProgress EventStatus = "IN_PROGRESS"
	EventDone       EventStatus = "DONE"
)

type LeagueEvent struct {
	EventID     int64       `db:"event_id"     json:"eventId"`
	LeagueID    int64       `db:"league_id"    json:"leagueId"`
	Status      EventStatus `db:"status"       json:"status"`
	Title       string      `db:"title"        json:"title"`
	StartDate   time.Time   `db:"start_date"   json:"startDate"`
	EndDate     time.Time   `db:"end_date"     json:"endDate"`
	Created     time.Time   `db:"created"      json:"created"`
	LastUpdated time.Time   `db:"last_updated" json:"lastUpdated"`
}
