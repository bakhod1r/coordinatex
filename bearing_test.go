package coordinatex

import "testing"

func TestBearing(t *testing.T) {
	o := Coordinate{0, 0}
	near(t, "north", Bearing(o, Coordinate{1, 0}).Degrees(), 0, 1e-9)
	near(t, "east", Bearing(o, Coordinate{0, 1}).Degrees(), 90, 1e-9)
	near(t, "south", Bearing(o, Coordinate{-1, 0}).Degrees(), 180, 1e-9)
	near(t, "west", Bearing(o, Coordinate{0, -1}).Degrees(), 270, 1e-9)
	near(t, "London-Paris", Bearing(london, paris).Degrees(), 148.1, 0.2)
}

func TestFinalBearing(t *testing.T) {
	near(t, "London-Paris final", FinalBearing(london, paris).Degrees(), 150.0, 0.3)
	o := Coordinate{0, 0}
	near(t, "equator east", FinalBearing(o, Coordinate{0, 10}).Degrees(), 90, 1e-9)
}

func TestDirection(t *testing.T) {
	o := Coordinate{0, 0}
	cases := map[Coordinate]CompassPoint{
		{1, 0}: North, {1, 1}: NorthEast, {0, 1}: East, {-1, 1}: SouthEast,
		{-1, 0}: South, {-1, -1}: SouthWest, {0, -1}: West, {1, -1}: NorthWest,
	}
	for to, want := range cases {
		if got := Direction(o, to); got != want {
			t.Errorf("Direction(%v) = %s, want %s", to, got, want)
		}
	}
	if CompassFromAngle(350) != North || CompassFromAngle(22.4) != North || CompassFromAngle(22.5) != NorthEast {
		t.Fatal("CompassFromAngle boundaries")
	}
}
