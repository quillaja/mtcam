package db

import (
	"math/rand"
	"testing"
	"time"

	"github.com/quillaja/mtcam/model"
)

const testConnection = "../new.db"

func TestMountains(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	mts, err := Mountains()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(mts)
}

func TestMountain(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	mt, err := Mountain(1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(mt)
}

func TestInsertMountain(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	m := model.Mountain{
		Name:        "TestPiss Peek",
		State:       "ASS",
		ElevationFt: rand.Intn(50000),
		Latitude:    1.0000,
		Longitude:   -10.000,
		TzLocation:  "America/New_York"}

	err = InsertMountain(&m)
	if err != nil {
		t.Fatal(err)
	}

	if m.ID == 0 {
		t.Fatal("id was 0")
	}
	t.Logf("new rowid = %d", m.ID)
}

func TestUpdateMountain(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	orig, err := Mountain(1)
	if err != nil {
		t.Fatal(err)
	}

	modified := orig
	modified.Name = "MODIFIED IN TEST"
	err = UpdateMountain(modified)
	if err != nil {
		t.Fatal(err)
	}

	modified, err = Mountain(modified.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(modified)

	// restore data, update modified to show change
	orig.Modified = time.Now()
	err = UpdateMountain(orig)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCameras(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	cams, err := Cameras()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(cams)
}

func TestCamera(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	cam, err := Camera(1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(cam)
}

func TestInsertCamera(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	c := model.Camera{
		Name:          "TEST PISS CAM",
		ElevationFt:   rand.Intn(3000),
		Latitude:      -99.9,
		Longitude:     69.69,
		Url:           "{{ http://piss.com }}",
		FileExtension: ".piss",
		IsActive:      true,
		Interval:      5,
		Delay:         20,
		Rules:         "{{ True }}",
		Comment:       "sucks",
		MountainID:    10}

	err = InsertCamera(&c)
	if err != nil {
		t.Fatal(err)
	}

	if c.ID == 0 {
		t.Fatal("id was 0")
	}
	t.Logf("new rowid = %d", c.ID)
}

func TestUpdateCamera(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	orig, err := Camera(1)
	if err != nil {
		t.Fatal(err)
	}

	modified := orig
	modified.Name = "MODIFIED IN TEST"
	err = UpdateCamera(modified)
	if err != nil {
		t.Fatal(err)
	}

	modified, err = Camera(modified.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(modified)

	// restore data. update modified to show change
	orig.Modified = time.Now()
	err = UpdateCamera(orig)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGroupCamerasByMountain(t *testing.T) {
	Connect(testConnection)
	defer Close()
	mts, _ := Mountains() // don't really need
	cams, _ := Cameras()
	mc := GroupCamerasByMountain(cams)
	for mId, cArr := range mc {
		t.Logf("%d - %s\n", mId, mts[mId].Name)
		for _, c := range cArr {
			t.Logf("\t%d - %s\n", c.ID, c.Name)
		}
	}
}

func TestScrapes(t *testing.T) {
	err := Connect(testConnection)
	defer Close()
	if err != nil {
		t.Fatal(err)
	}

	start := time.Date(2019, 6, 30, 10, 0, 0, 0, time.Local)
	scrapes, err := Scrapes(1, start, start.Add(5*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(scrapes)
}

func TestInsertScrape(t *testing.T) {
	t.SkipNow()

	Connect(testConnection)
	defer Close()

	s := model.Scrape{
		Created:  time.Now(),
		Result:   "TEST",
		Filename: "123abc.jpg",
		CameraID: 666}
	err := InsertScrape(&s)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", s)
}
