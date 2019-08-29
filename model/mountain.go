package model

import (
	"bytes"
	"strconv"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/quillaja/mtcam/astro"
)

type Mountain struct {
	ID          int            `json:"id"` // primary key
	Created     time.Time      `json:"-"`
	Modified    time.Time      `json:"-"`
	Name        string         `json:"name"`
	State       string         `json:"state"`
	ElevationFt int            `json:"elevation_ft"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	TzLocation  string         `json:"tz"`
	Pathname    string         `json:"pathname"`
	Cameras     map[int]Camera `json:"cams"`
}

type Camera struct {
	ID            int       `json:"id"` // primary key
	MountainID    int       `json:"-"`  // FK to mountain
	Created       time.Time `json:"-"`
	Modified      time.Time `json:"-"`
	Name          string    `json:"name"`
	ElevationFt   int       `json:"elevation_ft"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Comment       string    `json:"comment"`
	Interval      int       `json:"interval"`
	Delay         int       `json:"-"`
	FileExtension string    `json:"-"`
	Url           string    `json:"-"`         // template
	IsActive      bool      `json:"is_active"` // master on/off switch
	Rules         string    `json:"-"`         // template
	Pathname      string    `json:"pathname"`
}

func (c Camera) ExecuteUrl(data interface{}) (string, error) {

	funcs := template.FuncMap{
		"add":   func(i, j int) int { return i + j },
		"sub":   func(i, j int) int { return i - j },
		"mul":   func(i, j int) int { return i * j },
		"div":   func(i, j int) int { return i / j },
		"mod":   func(i, j int) int { return i % j },
		"floor": func(i, j int) int { return i - (i % j) },
	}

	t, err := template.New("url").Funcs(funcs).Parse(c.Url)
	if err != nil {
		return "", errors.Wrapf(err, "parsing camera url template (id=%d, name=%s)", c.ID, c.Name)
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		return "", errors.Wrapf(err, "executing camera url template (id=%d, name=%s)", c.ID, c.Name)
	}

	return buf.String(), nil
}

func (c Camera) ExecuteRules(data interface{}) (bool, error) {

	funcs := template.FuncMap{
		"add":   func(i, j int) int { return i + j },
		"sub":   func(i, j int) int { return i - j },
		"mul":   func(i, j int) int { return i * j },
		"div":   func(i, j int) int { return i / j },
		"mod":   func(i, j int) int { return i % j },
		"floor": func(i, j int) int { return i - (i % j) },

		"betweenRiseSet": func(now time.Time, sun astro.Data, hourOffset int) bool {
			offset := time.Duration(hourOffset)
			start := sun.SunTransit[astro.StartCivilTwilight].Add(-offset * time.Hour)
			end := sun.SunTransit[astro.EndCivilTwilight].Add(offset * time.Hour)
			return now.After(start) && now.Before(end)
		},

		"brightMoon": func(moon astro.Data) bool {
			switch moon.MoonPhase {
			case astro.FullMoon, astro.WaningGibbous, astro.WaxingGibbous:
				return true
			}
			return false
		},
	}

	t, err := template.New("rules").Funcs(funcs).Parse(c.Rules)
	if err != nil {
		return false, errors.Wrapf(err, "parsing camera rules template (id=%d, name=%s)", c.ID, c.Name)
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		return false, errors.Wrapf(err, "executing camera rules template (id=%d, name=%s)", c.ID, c.Name)
	}

	result, err := strconv.ParseBool(buf.String())
	if err != nil {
		return false, errors.Wrapf(err, "parsing rules result to bool (id=%d, name=%s)", c.ID, c.Name)
	}

	return result, nil
}

type Scrape struct {
	ID       int       `json:"-"` // primary key
	CameraID int       `json:"-"` // FK to camera
	Created  time.Time `json:"time"`
	Result   string    `json:"result"`
	Detail   string    `json:"detail"`
	Filename string    `json:"file"`
}

// Constants for Scrape.Result.
const (
	Success = "success"
	Failure = "failure"
	Idle    = "idle"
)
