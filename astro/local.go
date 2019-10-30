package astro

import (
	"time"

	"github.com/kelvins/sunrisesunset"
)

// GetLocal is like Get, but instead of querying the naval api, it uses
// a go pkg to calculate the sun rise and set. Only civil twilight and
// sun rise/set are provided.
func GetLocal(lat, lon float64, now time.Time) (Data, error) {
	_, offsetSec := now.Zone() // now's zone is same as mt's zone.
	offset := float64(offsetSec) / 3600
	rise, set, err := sunrisesunset.GetSunriseSunset(lat, lon, offset, now.UTC())
	if err != nil {
		return Data{}, err
	}
	rise = time.Date(now.Year(), now.Month(), now.Day(), rise.Hour(), rise.Minute(), rise.Second(), 0, now.Location())
	set = time.Date(now.Year(), now.Month(), now.Day(), set.Hour(), set.Minute(), set.Second(), 0, now.Location())

	sun := Data{
		Date: now,
		Lat:  lat,
		Lon:  lon,
		SunTransit: map[Phenom]time.Time{
			Rise:               rise,
			Set:                set,
			StartCivilTwilight: rise.Add(-30 * time.Minute), // civil twilight is about 30 mins before/after rise/set.
			EndCivilTwilight:   set.Add(30 * time.Minute),
		},
	}
	return sun, nil
}
