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

// Constants for moon phases.
// See: https://aa.usno.navy.mil/faq/docs/moon_phases.php
const (
	NewMoon        = "New Moon"
	WaxingCrescent = "Waxing Crescent"
	FirstQuarter   = "First Quarter"
	WaxingGibbous  = "Waxing Gibbous"
	FullMoon       = "Full Moon"
	WaningGibbous  = "Waning Gibbous"
	LastQuarter    = "Last Quarter"
	WaningCrescent = "Waning Crescent"
)

// Data is the astronomical information.
type Data struct {
	// The times of various transit phemonenon for the sun.
	SunTransit map[Phenom]time.Time
	// The times of various transit phemonenon for the moon.
	MoonTransit map[Phenom]time.Time
	// The moon's phase, such as "Last Quarter" or "Full Moon".
	MoonPhase string
	// The date for which the data applies. The 'time' portion
	// of the Date is irrelevant.
	Date time.Time
	// The location for the data.
	Lat, Lon float64
}

// a mapping of the string keys used in JSON to the
// constants used in this package.
// see: https://aa.usno.navy.mil/data/docs/api.php#rstt
var phenomKeys = map[string]Phenom{
	"BC": StartCivilTwilight,
	"R":  Rise,
	"U":  UpperTransit,
	"S":  Set,
	"EC": EndCivilTwilight,
	"L":  LowerTransit}
