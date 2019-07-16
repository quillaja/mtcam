// Package config opens and unmarshals a JSON config file. It also includes
// a function 'Watch' to be notified when a file changes on disk.
package config

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
)

// Read will open the JSON file at filename and attempt to unmarshal it
// into data.
func Read(filename string, data interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "attempting to read file")
	}

	err = json.Unmarshal(bytes, data)
	if err != nil {
		return errors.Wrap(err, "attempting to unmarshal data")
	}

	return nil
}

// Write will attempt to marshal data into JSON and write it to filename.
// 'data' should be a pointer, per standard json.Marshal() rules.
func Write(filename string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.Wrap(err, "attempting to marshal data")
	}

	err = ioutil.WriteFile(filename, bytes, 0666)
	if err != nil {
		return errors.Wrap(err, "attempting to write file")
	}

	return nil
}

// Watch checks filename for changes on disk, and calls action if
// the file has changed (error==nil) or if there is an error (error!=nil)
// reading the file's stats. Call cancel to stop watching the file.
func Watch(filename string, freq time.Duration, action func(string, error)) (cancel context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		oldstat, err := os.Lstat(filename)
		if err != nil {
			action(filename, err)
		}

		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(freq):
				newstat, err := os.Lstat(filename)
				if err != nil {
					action(filename, err)
					continue // go ahead and try again next round
				}

				if oldstat.ModTime() != newstat.ModTime() ||
					oldstat.Size() != newstat.Size() {
					action(filename, nil)
					oldstat = newstat
				}
			}
		}
	}()

	return
}
