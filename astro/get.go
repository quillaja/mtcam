// Package astro fetchest sun and moon data from the US Navy's
// "Astronomical Applications API".
//
// Website: https://aa.usno.navy.mil/data/docs/api.php
// API version: 2.2.1
package astro

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	urlfmt     = "https://api.usno.navy.mil/rstt/oneday?date=%s&coords=%f,%f&tz=%d&ID=%s"
	apikey     = "mtcam_v3"
	datelayout = "01/02/2006"
)

// Get queries the api for the sun and moon data for the date and location specified by
// date, lat, and lon. `date` must contain the correct `*time.Location` for the
// location, or incorrect data will be returned.
func Get(lat, lon float64, date time.Time) (Data, error) {
	// prepare request url
	_, offsetSec := date.Zone()
	offset := offsetSec / 3600 // convert sec to hours
	url := fmt.Sprintf(urlfmt, date.Format(datelayout), lat, lon, offset, apikey)

	// make request
	resp, err := http.Get(url)
	switch {
	case err != nil:
		return Data{}, errors.Wrap(err, "failed request to GET ")

	case resp.StatusCode != http.StatusOK:
		return Data{}, errors.New("request to GET " + url + " returned status code " + string(resp.StatusCode))
	}

	// read body
	buf, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to parse response body for "+url)
	}
	// parse json body
	rawdata := map[string]interface{}{}
	err = json.Unmarshal(buf, &rawdata)
	if err != nil {
		return Data{}, errors.Wrap(err, "failed to unmarshal json for "+url)
	}

	// extract valuable data from json
	data := Data{
		Date:        date,
		Lat:         lat,
		Lon:         lon,
		SunTransit:  map[Phenom]time.Time{},
		MoonTransit: map[Phenom]time.Time{}}
	extractData(&data, rawdata)

	return data, nil
}

// extractData is a function to pick through the JSON returned from
// the api request, stored in `raw`, and write the useful info into `data`.
func extractData(data *Data, raw map[string]interface{}) {

	date := data.Date

	// create helper func to extract transit phenomenon data
	transit := func(data map[Phenom]time.Time, transmap []interface{}) {
		const tlayout = "15:04" // time format that phenom are in (HH:MM)
		for _, p := range transmap {
			if phen, ok := p.(map[string]interface{}); ok {
				phenom := phenomKeys[phen["phen"].(string)]
				t, terr := time.Parse(tlayout, phen["time"].(string))
				if terr != nil {
					fmt.Println(terr)
					continue // skip if the time cannot be parsed
				}
				// build transit map combining date with phenom time
				data[phenom] = time.Date(
					date.Year(), date.Month(), date.Day(),
					t.Hour(), t.Minute(), 0, 0, date.Location())
			}
		}
	}

	// extract sun data
	if sundata, ok := raw["sundata"].([]interface{}); ok {
		transit(data.SunTransit, sundata)
	}

	// extract moon data
	if moondata, ok := raw["moondata"].([]interface{}); ok {
		transit(data.MoonTransit, moondata)
	}

	// extract moon phase
	// if "curphase" is provided use that. otherwise use "closestphase".
	// leave unchanged (should be empty string) if both fail
	if curphase, in := raw["curphase"]; in {
		data.MoonPhase = curphase.(string)
	} else if closest, ok := raw["closestphase"].(map[string]interface{}); ok {
		data.MoonPhase = closest["phase"].(string)
	}

}
