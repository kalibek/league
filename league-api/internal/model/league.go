package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type LeagueConfig struct {
	NumberOfAdvances int `json:"numberOfAdvances"`
	NumberOfRecedes  int `json:"numberOfRecedes"`
	GamesToWin       int `json:"gamesToWin"`
	GroupSize        int `json:"groupSize"`
	NumberOfTables   int `json:"numberOfTables"`
}

// Scan implements sql.Scanner so sqlx can read the JSONB column into LeagueConfig.
func (c *LeagueConfig) Scan(src interface{}) error {
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("LeagueConfig.Scan: unsupported type %T", src)
	}
	return json.Unmarshal(data, c)
}

// Value implements driver.Valuer so sqlx can write LeagueConfig as JSON.
func (c LeagueConfig) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type League struct {
	LeagueID    int64        `db:"league_id"     json:"leagueId"`
	Title       string       `db:"title"         json:"title"`
	Description string       `db:"description"   json:"description"`
	Config      LeagueConfig `db:"configuration" json:"configuration"`
	Created     time.Time    `db:"created"       json:"created"`
	LastUpdated time.Time    `db:"last_updated"  json:"lastUpdated"`
}

// LeagueWithStats extends League with event count and latest event date.
type LeagueWithStats struct {
	LeagueID        int64        `db:"league_id"          json:"leagueId"`
	Title           string       `db:"title"              json:"title"`
	Description     string       `db:"description"        json:"description"`
	Config          LeagueConfig `db:"configuration"      json:"configuration"`
	Created         time.Time    `db:"created"            json:"created"`
	LastUpdated     time.Time    `db:"last_updated"       json:"lastUpdated"`
	EventCount      int          `db:"event_count"        json:"eventCount"`
	LatestEventDate *time.Time   `db:"latest_event_date"  json:"latestEventDate"`
}

type LeagueMaintainer struct {
	UserID    int64  `db:"user_id"    json:"userId"`
	FirstName string `db:"first_name" json:"firstName"`
	LastName  string `db:"last_name"  json:"lastName"`
}

type LeagueSummary struct {
	LeagueWithStats
	Maintainers []LeagueMaintainer `json:"maintainers"`
}
