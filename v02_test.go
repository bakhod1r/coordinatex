package coordinatex

import (
	"math/rand"
	"testing"
	"time"
)

func TestContainsSpherical(t *testing.T) {
	if !square.ContainsSpherical(Coordinate{0.5, 0.5}) || square.ContainsSpherical(Coordinate{1.5, 0.5}) {
		t.Fatal("square")
	}
	// crosses the antimeridian: planar Contains gets this wrong
	dateline := Polygon{{-10, 170}, {10, 170}, {10, -170}, {-10, -170}}
	if !dateline.ContainsSpherical(Coordinate{0, 179}) || !dateline.ContainsSpherical(Coordinate{0, -179}) || dateline.ContainsSpherical(Coordinate{0, 0}) {
		t.Fatal("antimeridian")
	}
	// ring around the north pole
	polar := Polygon{{80, 0}, {80, 90}, {80, 180}, {80, -90}}
	if !polar.ContainsSpherical(Coordinate{85, 10}) || polar.ContainsSpherical(Coordinate{70, 10}) {
		t.Fatal("polar")
	}
	if !square.ContainsSpherical(Coordinate{0, 0}) {
		t.Fatal("vertex counts as inside")
	}
	if (Polygon{{0, 0}, {1, 1}}).ContainsSpherical(Coordinate{0.5, 0.5}) {
		t.Fatal("degenerate")
	}
	var _ Fence = SphericalPolygonFence{Polygon: dateline}
	if !(SphericalPolygonFence{Polygon: dateline}).Contains(Coordinate{0, 179}) {
		t.Fatal("fence")
	}
}

func TestPolygonDistanceTo(t *testing.T) {
	if square.DistanceTo(Coordinate{0.5, 0.5}) != 0 {
		t.Fatal("inside is 0")
	}
	near(t, "east of square", square.DistanceTo(Coordinate{0.5, 2}).Kilometers(), 111.18, 0.1)
	near(t, "corner", square.DistanceTo(Coordinate{-1, -1}).Kilometers(), DistanceBetween(Coordinate{-1, -1}, Coordinate{0, 0}).Kilometers(), 0.01)
	near(t, "fence", (PolygonFence{Polygon: square}).DistanceTo(Coordinate{0.5, 2}).Kilometers(), 111.18, 0.1)
	near(t, "circle fence", (CircleFence{Center: Coordinate{0, 0}, Radius: 10 * Kilometer}).DistanceTo(Coordinate{0, 1}).Kilometers(), 101.19, 0.05)
	near(t, "rect fence", (RectangleFence{Bounds: Bounds{0, 0, 1, 1}}).DistanceTo(Coordinate{0.5, 2}).Kilometers(), 111.18, 0.1)
	if (RectangleFence{Bounds: Bounds{0, 0, 1, 1}}).DistanceTo(Coordinate{0.5, 0.5}) != 0 {
		t.Fatal("rect inside")
	}
}

func TestCorridorAndCirclePolygon(t *testing.T) {
	route := Polyline{{0, 0}, {0, 1}, {1, 1}}
	f := CorridorFence{Route: route, Width: 1 * Kilometer}
	if !f.Contains(Coordinate{0.005, 0.5}) || f.Contains(Coordinate{0.02, 0.5}) || !f.Contains(Coordinate{0.5, 1.005}) {
		t.Fatal("corridor")
	}
	c := Circle{Center: tashkent, Radius: 5 * Kilometer}
	p := c.Polygon(64)
	if len(p) != 64 {
		t.Fatal(len(p))
	}
	for _, v := range p {
		near(t, "vertex on circle", DistanceBetween(tashkent, v).Meters(), 5000, 1e-6)
	}
	near(t, "area close to cap", p.Area().SquareKilometers(), c.Area().SquareKilometers(), c.Area().SquareKilometers()*0.01)
	if len(c.Polygon(1)) != 3 {
		t.Fatal("min vertices")
	}
}

func TestConvexHull(t *testing.T) {
	pts := []Coordinate{{0, 0}, {0, 2}, {2, 2}, {2, 0}, {1, 1}, {0.5, 1.5}, {0, 1}}
	h := ConvexHull(pts)
	if len(h) != 4 || h.IsClockwise() {
		t.Fatalf("%v", h)
	}
	for _, p := range pts {
		if !h.Contains(p) && square.DistanceTo(p) != 0 && !(p.Lat == 0 || p.Lng == 0 || p.Lat == 2 || p.Lng == 2) {
			t.Fatalf("%v outside hull", p)
		}
	}
	if !h.Contains(Coordinate{1, 1}) {
		t.Fatal("center")
	}
	if len(ConvexHull([]Coordinate{{0, 0}, {1, 1}})) != 2 || len(ConvexHull(nil)) != 0 {
		t.Fatal("small input")
	}
	if len(ConvexHull([]Coordinate{{0, 0}, {1, 1}, {2, 2}})) != 2 {
		t.Fatal("collinear")
	}
}

func TestDBSCAN(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	var pts []Coordinate
	for _, c := range []Coordinate{tashkent, samarkand} {
		for i := 0; i < 30; i++ {
			pts = append(pts, Destination(c, Distance(rng.Float64()*300), Angle(rng.Float64()*360)))
		}
	}
	pts = append(pts, london) // noise
	labels, n := DBSCAN(pts, 200*Meter, 4)
	if n != 2 {
		t.Fatalf("clusters = %d", n)
	}
	if labels[len(labels)-1] != Noise {
		t.Fatal("london should be noise")
	}
	if labels[0] == labels[30] || labels[0] == Noise || labels[30] == Noise {
		t.Fatal("two separate clusters", labels[0], labels[30])
	}
	for i := 1; i < 30; i++ {
		if labels[i] != labels[0] && labels[i] != Noise {
			t.Fatal("cluster split")
		}
	}
	if l, n := DBSCAN(nil, 1, 1); l != nil || n != 0 {
		t.Fatal("empty")
	}
}

func TestTrackStops(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	var tr Track
	add := func(c Coordinate, sec int) {
		tr = append(tr, TrackPoint{Coordinate: c, Time: t0.Add(time.Duration(sec) * time.Second)})
	}
	// drive, stop 10 min at A, drive, stop 2 min (too short), drive, stop 6 min at B
	sec := 0
	for i := 0; i < 5; i++ {
		add(Destination(tashkent, Distance(i*200), 90), sec)
		sec += 20
	}
	a := Destination(tashkent, 1000, 90)
	for i := 0; i <= 10; i++ {
		add(Destination(a, Distance(i%3*5), 0), sec)
		sec += 60
	}
	for i := 1; i <= 5; i++ {
		add(Destination(a, Distance(i*300), 90), sec)
		sec += 20
	}
	short := tr[len(tr)-1].Coordinate
	add(short, sec+60)
	add(short, sec+120)
	sec += 140
	b := Destination(short, 2000, 90)
	add(b, sec)
	for i := 1; i <= 6; i++ {
		add(b, sec+i*60)
	}
	stops := tr.Stops(50*Meter, 5*time.Minute)
	if len(stops) != 2 {
		t.Fatalf("stops = %d: %+v", len(stops), stops)
	}
	if !WithinRadius(stops[0].Center, a, 20) || stops[0].Duration() < 10*time.Minute || stops[0].Points != 11 {
		t.Fatalf("stop A %+v", stops[0])
	}
	if !WithinRadius(stops[1].Center, b, 1) {
		t.Fatalf("stop B %+v", stops[1])
	}
	trips := tr.Trips(50*Meter, 5*time.Minute)
	if len(trips) != 2 || trips[0][0].Time != t0 {
		t.Fatalf("trips = %d", len(trips))
	}
	if len((Track{}).Stops(1, time.Second)) != 0 {
		t.Fatal("empty")
	}
}
