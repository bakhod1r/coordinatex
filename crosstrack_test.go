package coordinatex

import "testing"

func TestCrossAlongTrack(t *testing.T) {
	a, b := Coordinate{0, 0}, Coordinate{0, 10}
	p := Coordinate{1, 5}
	near(t, "cross (left negative)", CrossTrackDistance(p, a, b).Kilometers(), -111.19, 0.05)
	near(t, "cross (right positive)", CrossTrackDistance(Coordinate{-1, 5}, a, b).Kilometers(), 111.19, 0.05)
	near(t, "along", AlongTrackDistance(p, a, b).Kilometers(), 555.9, 0.5)
	near(t, "along behind", AlongTrackDistance(Coordinate{0, -1}, a, b).Kilometers(), -111.19, 0.05)
	near(t, "on path", CrossTrackDistance(Coordinate{0, 3}, a, b).Meters(), 0, 1e-6)
}
