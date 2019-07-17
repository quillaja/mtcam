package db

import (
	"fmt"
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

func InsertMountain(m *model.Mountain) error {
	const query = `
	INSERT INTO mountain
		(created, modified, name, state,
		elevation_ft, latitude, longitude, tz_location)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?)`

	// ensure the user doesn't try to assign rowid
	if m.ID != 0 {
		return errors.Errorf("attempt to insert mountain with an existing ID (%d)", m.ID)
	}

	if m.Created.IsZero() {
		m.Created = time.Now()
	}
	if m.Modified.IsZero() {
		m.Modified = time.Now()
	}

	result, err := db.Exec(query,
		floorToSec(m.Created.In(time.UTC)), // ensure time in good format
		floorToSec(m.Modified.In(time.UTC)),
		m.Name,
		m.State,
		m.ElevationFt,
		m.Latitude,
		m.Longitude,
		m.TzLocation)
	if err != nil {
		return errors.Wrapf(err, "while inserting mountain (name: %s)", m.Name)
	}

	rowid, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "while getting new row id")
	}
	m.ID = int(rowid)

	return nil
}

func UpdateMountain(m model.Mountain) error {
	const query = `
	UPDATE mountain
	SET 
		modified = ?,
		name = ?,
		state = ?,
		elevation_ft = ?,
		latitude = ?,
		longitude = ?,
		tz_location = ?
	WHERE
		rowid=?`

	result, err := db.Exec(query,
		floorToSec(m.Modified.In(time.UTC)), // ensure time in good format
		m.Name,
		m.State,
		m.ElevationFt,
		m.Latitude,
		m.Longitude,
		m.TzLocation,
		m.ID)
	if err != nil {
		return errors.Wrapf(err, "updating mountain(id=%d)", m.ID)
	}
	if nrows, err := result.RowsAffected(); nrows != 1 {
		if err != nil {
			return errors.Wrap(err, "updating mountain(RowsAffected())")
		}
		return errors.New(
			fmt.Sprintf("%d rows affected. expected 1 when updating mountain(id=%d)", nrows, m.ID))
	}

	return nil
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

func CamerasOnMountain(mID int) (cams map[int]model.Camera, err error) {
	const query = `
	SELECT 
		rowid, created, modified, name,
		elevation_ft, latitude, longitude,
		url,
		file_ext, is_active, interval, delay, rules,
		comment,
		mountain_id 
	FROM 
		camera
	WHERE
		mountain_id=?`

	rows, err := db.Query(query, mID)
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

func InsertCamera(c *model.Camera) error {
	const query = `
	INSERT INTO camera
		(created, modified, name, elevation_ft, latitude, longitude,
		url, file_ext,
		is_active, interval, delay, rules,
		comment, mountain_id)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// ensure the user doesn't try to assign rowid
	if c.ID != 0 {
		return errors.Errorf("attempt to insert camera with an existing ID (%d)", c.ID)
	}

	if c.Created.IsZero() {
		c.Created = time.Now()
	}
	if c.Modified.IsZero() {
		c.Modified = time.Now()
	}

	result, err := db.Exec(query,
		floorToSec(c.Created.In(time.UTC)), // ensure time is in good format
		floorToSec(c.Modified.In(time.UTC)),
		c.Name,
		c.ElevationFt,
		c.Latitude,
		c.Longitude,
		c.Url,
		c.FileExtension,
		c.IsActive,
		c.Interval,
		c.Delay,
		c.Rules,
		c.Comment,
		c.MountainID)
	if err != nil {
		return errors.Wrapf(err, "while inserting cam (name: %s)", c.Name)
	}

	rowid, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "while getting new row id")
	}
	c.ID = int(rowid)

	return nil
}

func UpdateCamera(c model.Camera) error {
	const query = `
	UPDATE camera
	SET 
		modified = ?,
		name = ?,
		elevation_ft = ?,
		latitude = ?,
		longitude = ?,
		url = ?,
		file_ext = ?,
		is_active = ?,
		interval = ?,
		delay = ?,
		rules = ?,
		comment = ?,
		mountain_id = ?
	WHERE
		rowid=?`

	result, err := db.Exec(query,
		floorToSec(c.Modified.In(time.UTC)), // ensure time is in good format
		c.Name,
		c.ElevationFt,
		c.Latitude,
		c.Longitude,
		c.Url,
		c.FileExtension,
		c.IsActive,
		c.Interval,
		c.Delay,
		c.Rules,
		c.Comment,
		c.MountainID,
		c.ID)
	if err != nil {
		return errors.Wrapf(err, "updating camera(id=%d)", c.ID)
	}
	if nrows, err := result.RowsAffected(); nrows != 1 {
		if err != nil {
			return errors.Wrap(err, "updating camera(RowsAffected())")
		}
		return errors.New(
			fmt.Sprintf("%d rows affected. expected 1 when updating camera(id=%d)", nrows, c.ID))
	}

	return nil
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

	if s.Created.IsZero() {
		s.Created = time.Now()
	}

	result, err := db.Exec(query,
		floorToSec(s.Created.In(time.UTC)), // ensure time is in good format
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
