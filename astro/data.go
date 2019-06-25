package astro

import (
	"time"
)

// Phenom represents a transit phenomenon such as rising and setting.
type Phenom int

// The transit phenomenon.
const (
	StartCivilTwilight Phenom = iota
	Rise
	UpperTransit
	Set
	EndCivilTwilight
	LowerTransit
)

// Data is the astronomical information.
type Data struct {
	SunTransit  map[Phenom]time.Time
	MoonTransit map[Phenom]time.Time
	MoonPhase   string
	Date        time.Time
	Lat, Lon    float64
}

var phenomKeys = map[string]Phenom{
	"BC": StartCivilTwilight,
	"R":  Rise,
	"U":  UpperTransit,
	"S":  Set,
	"EC": EndCivilTwilight,
	"L":  LowerTransit}
