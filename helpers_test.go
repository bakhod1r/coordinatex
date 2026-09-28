package coordinatex

import (
	"math"
	"testing"
)

var (
	tashkent  = Coordinate{Lat: 41.2995, Lng: 69.2401}
	samarkand = Coordinate{Lat: 39.6542, Lng: 66.9597}
	london    = Coordinate{Lat: 51.5074, Lng: -0.1278}
	paris     = Coordinate{Lat: 48.8566, Lng: 2.3522}
)

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, want %v ±%v", name, got, want, tol)
	}
}

func nearCoord(t *testing.T, name string, got, want Coordinate, tol float64) {
	t.Helper()
	if !got.Equal(want, tol) {
		t.Errorf("%s = %v, want %v ±%v", name, got, want, tol)
	}
}
