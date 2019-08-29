package config

import (
	"testing"
	"time"
)

// TODO: write real tests

type testconfig struct {
	ImageDir string
	Database string
	Port     int
}

func TestWrite(t *testing.T) {
	cfg := testconfig{
		ImageDir: "img",
		Database: "something.db",
		Port:     666,
	}

	err := Write("test.json", cfg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRead(t *testing.T) {
	var cfg testconfig
	err := Read("test.json", &cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", cfg)
}

func TestWatch(t *testing.T) {
	cancel := Watch("test.json", 5*time.Second, func(f string, err error) {
		t.Log(f, err)
		if err != nil {
			t.Fatal(err)
		}
		var cfg testconfig
		Read(f, &cfg)
		t.Log(cfg)
	})
	defer cancel()

	time.Sleep(30 * time.Second)
}
