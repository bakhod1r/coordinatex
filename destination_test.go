package coordinatex

import "testing"

func TestDestination(t *testing.T) {
	d := Destination(tashkent, 10*Kilometer, 0)
	near(t, "north lng", d.Lng, tashkent.Lng, 1e-9)
	near(t, "north dist", DistanceBetween(tashkent, d).Meters(), 10000, 1e-3)
	if d.Lat <= tashkent.Lat {
		t.Fatal("expected north")
	}
	// round trip with bearing
	b := Bearing(london, paris)
	nearCoord(t, "London->Paris", Destination(london, DistanceBetween(london, paris), b), paris, 1e-6)
	// normalizes across antimeridian
	x := Destination(Coordinate{0, 179.9}, 50*Kilometer, 90)
	if x.Lng > 180 || x.Lng < -180 {
		t.Fatalf("not normalized: %v", x)
	}
}

func TestMidpoint(t *testing.T) {
	nearCoord(t, "equator", Midpoint(Coordinate{0, 0}, Coordinate{0, 90}), Coordinate{0, 45}, 1e-9)
	m := Midpoint(london, paris)
	near(t, "equal halves", DistanceBetween(london, m).Meters(), DistanceBetween(m, paris).Meters(), 1e-3)
}

func TestInterpolate(t *testing.T) {
	a, b := Coordinate{0, 0}, Coordinate{0, 90}
	nearCoord(t, "0", Interpolate(a, b, 0), a, 1e-9)
	nearCoord(t, "1", Interpolate(a, b, 1), b, 1e-9)
	nearCoord(t, "half", Interpolate(london, paris, 0.5), Midpoint(london, paris), 1e-9)
	q := Interpolate(london, paris, 0.25)
	near(t, "quarter", DistanceBetween(london, q).Meters(), DistanceBetween(london, paris).Meters()/4, 1e-3)
	nearCoord(t, "same point", Interpolate(a, a, 0.3), a, 1e-12)
}
