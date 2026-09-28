package coordinatex

import (
	"errors"
	"net/url"
	"testing"
	"time"
)

func TestMoreUnits(t *testing.T) {
	near(t, "Centimeters", (1 * Meter).Centimeters(), 100, 1e-12)
	near(t, "Yards", Yard.Yards(), 1, 1e-12)
	near(t, "Yard m", Yard.Meters(), 0.9144, 1e-12)
	a := Area(1e6)
	near(t, "km2", a.SquareKilometers(), 1, 1e-12)
	near(t, "ha", a.Hectares(), 100, 1e-9)
	near(t, "acres", Area(4046.8564224).Acres(), 1, 1e-9)
	near(t, "m2", a.SquareMeters(), 1e6, 0)
	if s := Area(12364e6).String(); s != "12364.00 km²" {
		t.Fatal(s)
	}
	if s := Area(50).String(); s != "50.00 m²" {
		t.Fatal(s)
	}
}

func TestAngleUtilities(t *testing.T) {
	near(t, "Diff wrap", float64(AngleDifference(350, 10)), 20, 1e-9)
	near(t, "Diff neg", float64(AngleDifference(10, 350)), -20, 1e-9)
	near(t, "Diff 180", float64(AngleDifference(0, 180)), 180, 1e-9)
	near(t, "Reverse", float64(Angle(30).Reverse()), 210, 1e-9)
	near(t, "Reverse wrap", float64(Angle(270).Reverse()), 90, 1e-9)
	if Compass4(Angle(44)) != North || Compass4(46) != East || Compass4(200) != South || Compass4(300) != West {
		t.Fatal("Compass4")
	}
	if Compass16(Angle(22.5)) != "NNE" || Compass16(0) != "N" || Compass16(348.75) != "N" || Compass16(337.5) != "NNW" || Compass16(180) != "S" {
		t.Fatal("Compass16")
	}
}

func TestCoordinateRoundText(t *testing.T) {
	nearCoord(t, "Round", Coordinate{41.311149, 69.279751}.Round(4), Coordinate{41.3111, 69.2798}, 1e-12)
	b, err := Coordinate{41.3111, 69.2797}.MarshalText()
	if err != nil || string(b) != "41.3111,69.2797" {
		t.Fatal(string(b), err)
	}
	var c Coordinate
	if err := c.UnmarshalText([]byte(`41°18'39.96"N 69°16'46.92"E`)); err != nil {
		t.Fatal(err)
	}
	nearCoord(t, "UnmarshalText", c, Coordinate{41.3111, 69.2797}, 1e-6)
	if err := c.UnmarshalText([]byte("nope")); err == nil {
		t.Fatal("bad text")
	}
	if _, err := (Coordinate{91, 0}).MarshalText(); err == nil {
		t.Fatal("invalid text")
	}
}

func TestIntermediatePoints(t *testing.T) {
	pts := IntermediatePoints(Coordinate{0, 0}, Coordinate{0, 10}, 4)
	if len(pts) != 6 {
		t.Fatal(len(pts))
	}
	nearCoord(t, "first", pts[0], Coordinate{0, 0}, 1e-9)
	nearCoord(t, "mid", pts[1], Coordinate{0, 2}, 1e-9)
	nearCoord(t, "last", pts[5], Coordinate{0, 10}, 1e-9)

	every := PointsEvery(Coordinate{0, 0}, Coordinate{0, 1}, 50*Kilometer)
	// 111.19 km: 0, 50, 100, 111.19
	if len(every) != 4 {
		t.Fatal(len(every))
	}
	near(t, "step", DistanceBetween(every[0], every[1]).Kilometers(), 50, 1e-6)
	nearCoord(t, "end", every[3], Coordinate{0, 1}, 1e-12)
	if got := PointsEvery(tashkent, samarkand, 0); len(got) != 2 {
		t.Fatal("zero step", got)
	}
}

func TestDistanceMatrix(t *testing.T) {
	m := DistanceMatrix([]Coordinate{tashkent, london}, []Coordinate{samarkand, paris, tashkent})
	if len(m) != 2 || len(m[0]) != 3 {
		t.Fatal("shape")
	}
	near(t, "m[0][0]", m[0][0].Meters(), DistanceBetween(tashkent, samarkand).Meters(), 0)
	near(t, "m[1][1]", m[1][1].Meters(), DistanceBetween(london, paris).Meters(), 0)
	near(t, "m[0][2]", m[0][2].Meters(), 0, 0)
}

func TestBoundsExtras(t *testing.T) {
	a := Bounds{0, 0, 10, 10}
	got, ok := a.Intersection(Bounds{5, 5, 15, 15})
	if !ok || got != (Bounds{5, 5, 10, 10}) {
		t.Fatal(got, ok)
	}
	if _, ok := a.Intersection(Bounds{20, 20, 30, 30}); ok {
		t.Fatal("disjoint")
	}
	one := Bounds{0, 0, 1, 1}
	near(t, "Area", one.Area().SquareKilometers(), 12364, 5)
	near(t, "Perimeter", one.Perimeter().Kilometers(), 4*111.19, 0.5)
	parts := a.Split(2, 2)
	if len(parts) != 4 || parts[0] != (Bounds{0, 0, 5, 5}) || parts[3] != (Bounds{5, 5, 10, 10}) {
		t.Fatal(parts)
	}
	if a.Split(0, 1) != nil {
		t.Fatal("bad split")
	}
	x := Bounds{-10, 170, 10, -170}.Split(1, 2)
	if len(x) != 2 || x[0].MinLng != 170 || x[1].MaxLng != -170 || x[1].MinLng != -180 {
		t.Fatal("antimeridian split", x)
	}
}

func TestGeohashExtras(t *testing.T) {
	if p, err := GeohashParent("u4pru"); err != nil || p != "u4pr" {
		t.Fatal(p, err)
	}
	if _, err := GeohashParent("u"); !errors.Is(err, ErrInvalidGeohash) {
		t.Fatal(err)
	}
	ch, err := GeohashChildren("u4pr")
	if err != nil || len(ch) != 32 || ch[0] != "u4pr0" || ch[31] != "u4prz" {
		t.Fatal(ch, err)
	}
	if _, err := GeohashChildren("u4pruydqqvjk"); !errors.Is(err, ErrInvalidGeohash) {
		t.Fatal("max precision children", err)
	}
	if !IsValidGeohash("u4pru") || IsValidGeohash("u4pa") || IsValidGeohash("") {
		t.Fatal("IsValidGeohash")
	}
	cells := GeohashesAround(tashkent, 2*Kilometer)
	if len(cells) != 9 {
		t.Fatal(cells)
	}
	// every point within radius is in one of the cells
	set := map[string]bool{}
	for _, c := range cells {
		set[c] = true
	}
	p := len(cells[0])
	for brg := Angle(0); brg < 360; brg += 15 {
		if h := EncodeGeohash(Destination(tashkent, 2*Kilometer, brg), p); !set[h] {
			t.Fatalf("bearing %v not covered (%s)", brg, h)
		}
	}
}

func TestCircle(t *testing.T) {
	c := Circle{Center: Coordinate{0, 0}, Radius: 10 * Kilometer}
	if !c.Contains(Coordinate{0, 0.05}) || c.Contains(Coordinate{0, 1}) {
		t.Fatal("Contains")
	}
	near(t, "Area", c.Area().SquareKilometers(), 314.16, 0.1)
	near(t, "Circumference", c.Circumference().Kilometers(), 62.83, 0.01)
	near(t, "DistanceTo outside", c.DistanceTo(Coordinate{0, 1}).Kilometers(), 101.19, 0.05)
	near(t, "DistanceTo inside", c.DistanceTo(Coordinate{0, 0}).Kilometers(), -10, 1e-9)
	if !c.Bounds().Contains(Destination(c.Center, 9.99*Kilometer, 45)) {
		t.Fatal("Bounds")
	}
	d := Circle{Center: Coordinate{0, 0.15}, Radius: 10 * Kilometer}
	if !c.Intersects(d) || c.Intersects(Circle{Center: Coordinate{0, 1}, Radius: 1}) {
		t.Fatal("Intersects")
	}
	var _ Fence = c
}

func TestPolygonIsSimple(t *testing.T) {
	if !square.IsSimple() {
		t.Fatal("square is simple")
	}
	bowtie := Polygon{{0, 0}, {1, 1}, {1, 0}, {0, 1}}
	if bowtie.IsSimple() {
		t.Fatal("bowtie self-intersects")
	}
}

func TestPolylineExtras(t *testing.T) {
	l := Polyline{{0, 0}, {0, 1}, {0, 3}}
	r := l.Reverse()
	if r[0] != l[2] || r[2] != l[0] || l[0] != (Coordinate{0, 0}) {
		t.Fatal("Reverse", r)
	}
	seg := l.SegmentLengths()
	if len(seg) != 2 || seg[1] <= seg[0] {
		t.Fatal(seg)
	}
	br := l.Bearings()
	if len(br) != 2 {
		t.Fatal(br)
	}
	near(t, "bearing", br[0].Degrees(), 90, 1e-9)
	p, ok := l.PointAt(l.Length() / 2)
	if !ok {
		t.Fatal("PointAt")
	}
	nearCoord(t, "PointAt mid", p, Coordinate{0, 1.5}, 1e-9)
	p, _ = l.PointAt(-5)
	nearCoord(t, "PointAt clamp start", p, l[0], 0)
	p, _ = l.PointAt(1e12)
	nearCoord(t, "PointAt clamp end", p, l[2], 0)
	if _, ok := (Polyline{}).PointAt(1); ok {
		t.Fatal("empty")
	}
	if l.Deviates(Coordinate{0.001, 2}, 500*Meter) || !l.Deviates(Coordinate{0.1, 2}, 500*Meter) {
		t.Fatal("Deviates")
	}
	if !(Polyline{}).Deviates(tashkent, 1) {
		t.Fatal("empty route deviates")
	}
}

func TestFenceTransition(t *testing.T) {
	f := CircleFence{Center: tashkent, Radius: 1 * Kilometer}
	far := Destination(tashkent, 5*Kilometer, 0)
	cases := []struct {
		prev, cur Coordinate
		want      Transition
	}{
		{far, tashkent, Entered},
		{tashkent, far, Exited},
		{tashkent, tashkent, StayedInside},
		{far, far, StayedOutside},
	}
	for _, c := range cases {
		if got := DetectTransition(f, c.prev, c.cur); got != c.want {
			t.Errorf("got %v want %v", got, c.want)
		}
	}
	if Entered.String() != "entered" || Transition(99).String() != "unknown" {
		t.Fatal("String")
	}
}

func TestTrack(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	a := Coordinate{0, 0}
	b := Destination(a, 1*Kilometer, 90)
	c := Destination(b, 2*Kilometer, 90)
	tr := Track{
		{Coordinate: a, Time: t0},
		{Coordinate: b, Time: t0.Add(100 * time.Second)}, // 10 m/s
		{Coordinate: c, Time: t0.Add(200 * time.Second)}, // 20 m/s
	}
	near(t, "Distance", tr.Distance().Meters(), 3000, 1e-6)
	if tr.Duration() != 200*time.Second {
		t.Fatal(tr.Duration())
	}
	near(t, "Avg", tr.AverageSpeed().MetersPerSecond(), 15, 1e-6)
	near(t, "Max", tr.MaxSpeed().MetersPerSecond(), 20, 1e-6)
	near(t, "KMH", Speed(10).KilometersPerHour(), 36, 1e-9)
	near(t, "MPH", Speed(1).MilesPerHour(), 2.2369362920544, 1e-9)
	near(t, "Knots", Speed(1).Knots(), 1.9438444924406, 1e-9)
	if len(tr.Speeds()) != 2 || (Track{}).AverageSpeed() != 0 || (Track{}).Duration() != 0 {
		t.Fatal("edge")
	}
	if len(tr.Polyline()) != 3 {
		t.Fatal("Polyline")
	}

	// spike: point 50 km away after 1 s is noise
	noisy := Track{
		{Coordinate: a, Time: t0},
		{Coordinate: Destination(a, 50*Kilometer, 0), Time: t0.Add(time.Second)},
		{Coordinate: b, Time: t0.Add(100 * time.Second)},
		{Coordinate: b, Time: t0.Add(100 * time.Second)}, // zero dt duplicate is dropped
	}
	f := noisy.FilterSpeed(Speed(50))
	if len(f) != 2 || f[1].Coordinate != b {
		t.Fatal("FilterSpeed", f)
	}
	if !(Track{{Coordinate: a}, {Coordinate: Destination(a, 5, 0)}}).IsStationary(10*Meter) || tr.IsStationary(10*Meter) {
		t.Fatal("IsStationary")
	}
	if !(Track{}).IsStationary(1) {
		t.Fatal("empty stationary")
	}
}

func TestQueryHelpers(t *testing.T) {
	q := url.Values{"lat": {"41.3111"}, "lng": {"69.2797"}, "radius": {"5000"}, "bbox": {"68,40,70,42"}}
	c, err := CoordinateFromQuery(q, "lat", "lng")
	if err != nil || c != (Coordinate{41.3111, 69.2797}) {
		t.Fatal(c, err)
	}
	r, err := RadiusFromQuery(q, "radius", 10*Kilometer)
	if err != nil || r != 5*Kilometer {
		t.Fatal(r, err)
	}
	if _, err := RadiusFromQuery(url.Values{"radius": {"50000"}}, "radius", 10*Kilometer); err == nil {
		t.Fatal("radius over max")
	}
	if _, err := RadiusFromQuery(url.Values{"radius": {"-1"}}, "radius", 10*Kilometer); err == nil {
		t.Fatal("negative radius")
	}
	if _, err := RadiusFromQuery(url.Values{}, "radius", 10*Kilometer); err == nil {
		t.Fatal("missing radius")
	}
	b, err := BoundsFromQuery(q, "bbox")
	if err != nil || b != (Bounds{MinLat: 40, MinLng: 68, MaxLat: 42, MaxLng: 70}) {
		t.Fatal(b, err)
	}
	for _, bad := range []url.Values{{"lat": {"x"}, "lng": {"1"}}, {"lat": {"1"}}, {"lat": {"91"}, "lng": {"0"}}} {
		if _, err := CoordinateFromQuery(bad, "lat", "lng"); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
	for _, bad := range []string{"", "1,2,3", "a,b,c,d", "0,10,1,5"} {
		if _, err := BoundsFromQuery(url.Values{"bbox": {bad}}, "bbox"); err == nil {
			t.Errorf("bbox %q accepted", bad)
		}
	}
}
