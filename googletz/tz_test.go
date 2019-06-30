package googletz

import (
	"reflect"
	"testing"
)

// TODO: api key
const testkey = "YOUR_API_KEY"

// TODO: this doesn't fully test the result, just the "Id" field.
func TestGet(t *testing.T) {
	type args struct {
		lat    float64
		lon    float64
		apikey string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "mt_hood",
			args: args{lat: 45.373439, lon: -121.695962, apikey: testkey},
			want: "America/Los_Angeles",
		},
		{
			name: "mt_blanc",
			args: args{lat: 45.833611, lon: 6.865, apikey: testkey},
			want: "Europe/Paris",
		},
		{
			// invalid api key
			name:    "mt_hood_apikey_error",
			args:    args{lat: 45.373439, lon: -121.695962, apikey: "fakekey"},
			want:    "", // value of Id for "zero" Timezone
			wantErr: true,
		},
		{
			// location in ocean off oregon coast
			name:    "bad_location_error",
			args:    args{lat: 45.5, lon: -125.0, apikey: testkey},
			want:    "", // value of Id for "zero" Timezone
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.lat, tt.args.lon, tt.args.apikey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v | got: %v", err, tt.wantErr, got)
				return
			}
			if !reflect.DeepEqual(got.Id, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
