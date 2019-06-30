// Package googletz implements a simple api for accessing the Google
// Timezone API.
// See: https://developers.google.com/maps/documentation/timezone/intro
package googletz

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// api endpoint
const urlfmt = "https://maps.googleapis.com/maps/api/timezone/json?location=%f,%f&timestamp=%d&key=%s"

// Status codes that can be returned by the API.
const (
	StatusOK             = "OK"
	StatusInvalidRequest = "INVALID_REQUEST"
	StatusOverDailyLimit = "OVER_DAILY_LIMIT"
	StatusOverQueryLimit = "OVER_QUERY_LIMIT"
	StatusRequestDenied  = "REQUEST_DENIED"
	StatusUnknownError   = "UNKNOWN_ERROR"
	StatusZeroResults    = "ZERO_RESULTS"
)

// Timezone holds data returned from the Google Timezone API.
type Timezone struct {
	DSTOffset    int    `json:"dstOffset"`
	RawOffset    int    `json:"rawOffset"`
	Id           string `json:"timeZoneId"`
	Name         string `json:"timeZoneName"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

// Get gets current the timezone info from the Google Timezone API.
// lat and lon specify the location, which must be on land.
// apikey is the API key you get from Google.
func Get(lat, lon float64, apikey string) (Timezone, error) {
	return GetAt(lat, lon, time.Now(), apikey)
}

// GetAt gets the timezone info from the Google Timezone API.
// lat and lon specify the location, which must be on land.
// t is the point in time at which to request the timezone info.
// apikey is the API key you get from Google.
func GetAt(lat, lon float64, t time.Time, apikey string) (tz Timezone, err error) {
	// timestamp is seconds since midnight, January 1, 1970 UTC.
	url := fmt.Sprintf(urlfmt, lat, lon, t.Unix(), apikey)
	resp, err := http.Get(url)
	if err != nil {
		return tz, errors.Wrap(err, "request to google tz api failed")
	}
	if resp.StatusCode != http.StatusOK {
		return tz, errors.New("request to google tz api returned status code " + string(resp.StatusCode))
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return tz, errors.Wrap(err, "while reading response body")
	}

	err = json.Unmarshal(data, &tz) // finally actually do the thing
	switch {
	case err != nil:
		err = errors.Wrap(err, "while unmarshalling json")

	case tz.Status != StatusOK:
		err = errors.Errorf("request succeeded but google returned error: (%s: %s)", tz.Status, tz.ErrorMessage)
	}

	return
}
