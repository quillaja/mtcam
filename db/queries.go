package db

import (
	"time"

	"github.com/pkg/errors"
	"github.com/quillaja/mtcam/model"
)

func Mountains() (mts map[int]model.Mountain, err error) {
	const query = `
	SELECT 
		rowid, created, modified, 
		name, state, 
		elevation_ft, latitude, longitude, tz_location 
	FROM 
		mountain`

	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Mountains()")
	}
	defer rows.Close()

	mts = make(map[int]model.Mountain)
	var mt model.Mountain
	for rows.Next() {
		err2 := rows.Scan(
			&mt.ID,
			&mt.Created,
			&mt.Modified,
			&mt.Name,
			&mt.State,
			&mt.ElevationFt,
			&mt.Latitude,
			&mt.Longitude,
			&mt.TzLocation)
		if err2 != nil {
			// TODO: something with the error
		}
		mts[mt.ID] = mt
	}

	return
}

func Mountain(id int) (m model.Mountain, err error) {
	const query = `
	SELECT rowid, created, modified, name, state, elevation_ft, latitude, longitude, tz_location
	FROM mountain
	WHERE
		rowid=?
	LIMIT 1`

	row := db.QueryRow(query, id)
	err = row.Scan(
		&m.ID,
		&m.Created,
		&m.Modified,
		&m.Name,
		&m.State,
		&m.ElevationFt,
		&m.Latitude,
		&m.Longitude,
		&m.TzLocation)
	if err != nil {
		return m, errors.Wrap(err, "db.Mountain(id)")
	}

	return
}

func Cameras() (cams map[int]model.Camera, err error) {
	const query = `
	SELECT 
		rowid, created, modified, name,
		elevation_ft, latitude, longitude,
		url,
		file_ext, is_active, interval, delay, rules,
		comment,
		mountain_id 
	FROM 
		camera`

	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Camera()")
	}
	defer rows.Close()

	cams = make(map[int]model.Camera)
	var cam model.Camera
	for rows.Next() {
		err2 := rows.Scan(
			&cam.ID,
			&cam.Created,
			&cam.Modified,
			&cam.Name,
			&cam.ElevationFt,
			&cam.Latitude,
			&cam.Longitude,
			&cam.Url,
			&cam.FileExtension,
			&cam.IsActive,
			&cam.Interval,
			&cam.Delay,
			&cam.Rules,
			&cam.Comment,
			&cam.MountainID)
		if err2 != nil {
			// TODO: something with the error
		}
		cams[cam.ID] = cam
	}

	return
}

func GroupCamerasByMountain(cams map[int]model.Camera) (groups map[int][]model.Camera) {
	groups = make(map[int][]model.Camera)
	for _, c := range cams {
		groups[c.MountainID] = append(groups[c.MountainID], c)
	}
	return
}

func Camera(id int) (c model.Camera, err error) {
	const query = `
	SELECT
		rowid, created, modified, name,
		elevation_ft, latitude, longitude,
		url, file_ext,
		is_active, interval, delay, rules,
		comment, mountain_id
	FROM camera
	WHERE
		rowid=?
	LIMIT 1`

	row := db.QueryRow(query, id)
	err = row.Scan(
		&c.ID,
		&c.Created,
		&c.Modified,
		&c.Name,
		&c.ElevationFt,
		&c.Latitude,
		&c.Longitude,
		&c.Url,
		&c.FileExtension,
		&c.IsActive,
		&c.Interval,
		&c.Delay,
		&c.Rules,
		&c.Comment,
		&c.MountainID)
	if err != nil {
		return c, errors.Wrap(err, "db.Camera(id)")
	}

	return
}

func Scrapes(camID int, start, end time.Time) (scrapes []model.Scrape, err error) {
	const query = `
	SELECT rowid, created, result, detail, filename, camera_id
	FROM scrape
	WHERE
		camera_id=?
		AND
		created BETWEEN ? AND ?
	ORDER BY
		created ASC`

	rows, err := db.Query(query, camID, start, end)
	if err != nil {
		return nil, errors.Wrap(err, "db.Scrapes()")
	}
	defer rows.Close()

	scrapes = make([]model.Scrape, 0)
	var s model.Scrape
	for rows.Next() {
		err2 := rows.Scan(
			&s.ID,
			&s.Created,
			&s.Result,
			&s.Detail,
			&s.Filename,
			&s.CameraID)
		// TODO: no longer needed because all tables converted to contain tz info
		// s.Created = time.Date(s.Created.Year(), s.Created.Month(), s.Created.Day(),
		// 	s.Created.Hour(), s.Created.Minute(), s.Created.Second(), s.Created.Nanosecond(),
		// 	time.Local)
		if err2 != nil {
			// TODO: something with the error
		}
		scrapes = append(scrapes, s)
	}

	return
}

func MostRecentScrape(camID int) (s model.Scrape, err error) {
	const query = `
	SELECT rowid, created, result, detail, filename, camera_id
	FROM scrape
	WHERE
		camera_id=?
	ORDER BY
		created DESC
	LIMIT 1`

	row := db.QueryRow(query, camID)
	err = row.Scan(
		&s.ID,
		&s.Created,
		&s.Result,
		&s.Detail,
		&s.Filename,
		&s.CameraID)
	if err != nil {
		return s, errors.Wrap(err, "db.MostRecentScrape()")
	}

	return
}

func InsertScrape(s *model.Scrape) error {
	const query = `
	INSERT INTO scrape
		(created, result, detail, filename, camera_id)
	VALUES
		(?, ?, ?, ?, ?)`

	// ensure the user doesn't try to assign rowid
	if s.ID != 0 {
		return errors.Errorf("attempt to insert scrape with an existing ID (%d)", s.ID)
	}

	result, err := db.Exec(query,
		floorToSec(s.Created.In(time.UTC)),
		s.Result,
		s.Detail,
		s.Filename,
		s.CameraID)
	if err != nil {
		return errors.Wrapf(err, "while inserting scrape (cam: %d, time: %s)",
			s.CameraID, s.Created.Format(time.RFC3339))
	}

	rowid, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "while getting new row id")
	}
	s.ID = int(rowid)

	return nil
}

// floorToSec zeros the nanosecond component of a time.
func floorToSec(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}
