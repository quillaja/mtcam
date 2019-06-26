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
var phenomKeys = map[string]Phenom{
	"BC": StartCivilTwilight,
	"R":  Rise,
	"U":  UpperTransit,
	"S":  Set,
	"EC": EndCivilTwilight,
	"L":  LowerTransit}
