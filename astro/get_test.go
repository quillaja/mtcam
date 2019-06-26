package astro

import (
	"testing"
	"time"
)

// TODO: write a real test
func TestGet(t *testing.T) {
	pacificTZ, _ := time.LoadLocation("America/Los_Angeles")
	japanTZ, _ := time.LoadLocation("Asia/Tokyo")

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
		{name: "mt fuji 6/25/19", args: args{
			lat: 35.360388, lon: 138.727724,
			date: time.Date(2019, 6, 25, 0, 0, 0, 0, japanTZ)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.lat, tt.args.lon, tt.args.date)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%+v\n", got)
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
