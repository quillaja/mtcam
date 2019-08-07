package model

import (
	"bytes"
	"strconv"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

type Mountain struct {
	ID                  int // primary key
	Created, Modified   time.Time
	Name                string
	State               string
	ElevationFt         int
	Latitude, Longitude float64
	TzLocation          string
	Pathname            string
}

type Camera struct {
	ID                  int // primary key
	MountainID          int // FK to mountain
	Created, Modified   time.Time
	Name                string
	ElevationFt         int
	Latitude, Longitude float64
	Comment             string
	Interval, Delay     int
	FileExtension       string
	Url                 string // template
	IsActive            bool   // master on/off switch
	Rules               string // template
	Pathname            string
}

func (c Camera) ExecuteUrl(data interface{}) (string, error) {

	t, err := template.New("url").Parse(c.Url)
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

	t, err := template.New("rules").Parse(c.Rules)
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
	ID       int // primary key
	CameraID int // FK to camera
	Created  time.Time
	Result   string
	Detail   string
	Filename string
}

// Constants for Scrape.Result.
const (
	Success = "success"
	Failure = "failure"
	Idle    = "idle"
)
