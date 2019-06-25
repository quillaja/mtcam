// Package astro fetchest sun and moon data from the US Navy's
// "Astronomical Applications API".
//
// Website: https://aa.usno.navy.mil/data/docs/api.php
// API version: 2.2.1
package astro

import (
	"testing"
	"time"
)

// TODO: write a real test
func TestGet(t *testing.T) {
	pacificTZ, _ := time.LoadLocation("America/Los_Angeles")

	type args struct {
		lat  float64
		lon  float64
		date time.Time
	}
	tests := []struct {
		name string
		args args
		// want    Data
		// wantErr bool
	}{
		{name: "mt hood 6/24/19", args: args{
			lat: 45.34511, lon: -121.711769,
			date: time.Date(2019, 6, 24, 0, 0, 0, 0, pacificTZ)},
		},
		{name: "mt hood 6/25/19", args: args{
			lat: 45.34511, lon: -121.711769,
			date: time.Date(2019, 6, 25, 0, 0, 0, 0, pacificTZ)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.lat, tt.args.lon, tt.args.date)
			t.Logf("%+v\n", got)
			if err != nil {
				t.Fatal(err)
			}
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			// 	return
			// }
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Get() = %v, want %v", got, tt.want)
			// }
		})
	}
}
