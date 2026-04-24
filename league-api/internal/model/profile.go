package model

import "time"

type Country struct {
	CountryID int    `db:"country_id" json:"countryId"`
	Name      string `db:"name"       json:"name"`
	Code      string `db:"code"       json:"code"`
}

type City struct {
	CityID    int    `db:"city_id"    json:"cityId"`
	Name      string `db:"name"       json:"name"`
	CountryID int    `db:"country_id" json:"countryId"`
}

type Blade struct {
	BladeID int    `db:"blade_id" json:"bladeId"`
	Name    string `db:"name"     json:"name"`
}

type Rubber struct {
	RubberID int    `db:"rubber_id" json:"rubberId"`
	Name     string `db:"name"      json:"name"`
}

type PlayerProfileRow struct {
	ProfileID   int64      `db:"profile_id"`
	UserID      int64      `db:"user_id"`
	CountryID   *int       `db:"country_id"`
	CityID      *int       `db:"city_id"`
	Birthdate   *time.Time `db:"birthdate"`
	Grip        *string    `db:"grip"`
	Gender      *string    `db:"gender"`
	BladeID     *int       `db:"blade_id"`
	FhRubberID  *int       `db:"fh_rubber_id"`
	BhRubberID  *int       `db:"bh_rubber_id"`
	Created     time.Time  `db:"created"`
	LastUpdated time.Time  `db:"last_updated"`
}

// PlayerProfileDetail is the full enriched profile returned by the API.
type PlayerProfileDetail struct {
	UserID      int64      `json:"userId"`
	FirstName   string     `json:"firstName"`
	LastName    string     `json:"lastName"`
	Country     *Country   `json:"country"`
	City        *City      `json:"city"`
	Birthdate   *time.Time `json:"birthdate"`
	Grip        *string    `json:"grip"`
	Gender      *string    `json:"gender"`
	Blade       *Blade     `json:"blade"`
	FhRubber    *Rubber    `json:"fhRubber"`
	BhRubber    *Rubber    `json:"bhRubber"`
	IsComplete  bool       `json:"isComplete"`
}
