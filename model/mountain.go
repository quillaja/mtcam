package model

import "time"

type Mountain struct {
	ID                  int // primary key
	Created, Modified   time.Time
	Name                string
	State               string
	ElevationFt         float64
	Latitude, Longitude float64
	TzLocation          string
}

type Camera struct {
	ID                  int // primary key
	MountainID          int // FK to mountain
	Created, Modified   time.Time
	Name                string
	ElevationFt         float64
	Latitude, Longitude float64
	Comment             string
	Interval, Delay     int
	FileExtension       string
	Url                 string // template
	IsActive            bool   // master on/off switch
	Rules               string // template
}

type Scrape struct {
	ID       int // primary key
	CameraID int // FK to camera
	Created  time.Time
	Result   string
	Detail   string
	Filename string
}
