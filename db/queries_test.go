package db

import (
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
