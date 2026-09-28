package coordinatex

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

var (
	flinders  = Coordinate{-(37 + 57/60.0 + 3.72030/3600), 144 + 25/60.0 + 29.52440/3600}
	buninyong = Coordinate{-(37 + 39/60.0 + 10.15610/3600), 143 + 55/60.0 + 35.38390/3600}
)

func TestDestinationVincenty(t *testing.T) {
	brg := Angle(306 + 52/60.0 + 5.37/3600)
	got, err := DestinationVincenty(flinders, 54972.271, brg)
	if err != nil {
		t.Fatal(err)
	}
	nearCoord(t, "direct", got, buninyong, 1e-8)
	got, _ = DestinationVincenty(tashkent, 0, 0)
	nearCoord(t, "zero", got, tashkent, 1e-12)
	x, _ := DestinationVincenty(Coordinate{0, 179.9}, 50*Kilometer, 90)
	if x.Lng > 180 || x.Lng < -180 {
		t.Fatal("not wrapped", x)
	}
}

func TestSegmentIntersection(t *testing.T) {
	p, ok := SegmentIntersection(Coordinate{0, -10}, Coordinate{0, 10}, Coordinate{-10, 0}, Coordinate{10, 0})
	if !ok {
		t.Fatal("expected intersection")
	}
	nearCoord(t, "cross", p, Coordinate{0, 0}, 1e-9)
	if _, ok := SegmentIntersection(Coordinate{0, -10}, Coordinate{0, 10}, Coordinate{5, 20}, Coordinate{10, 20}); ok {
		t.Fatal("disjoint")
	}
	// great circles cross but not within the short segments
	if _, ok := SegmentIntersection(Coordinate{0, 1}, Coordinate{0, 10}, Coordinate{-10, 0}, Coordinate{10, 0}); ok {
		t.Fatal("outside arc")
	}
	// antimeridian
	p, ok = SegmentIntersection(Coordinate{0, 170}, Coordinate{0, -170}, Coordinate{-5, 180}, Coordinate{5, 180})
	if !ok || math.Abs(math.Abs(p.Lng)-180) > 1e-9 {
		t.Fatal("antimeridian", p, ok)
	}
	if _, ok := SegmentIntersection(Coordinate{0, 0}, Coordinate{0, 10}, Coordinate{0, 2}, Coordinate{0, 5}); ok {
		t.Fatal("collinear reported as single point")
	}
}

func TestPolygonRelations(t *testing.T) {
	big := Polygon{{0, 0}, {0, 10}, {10, 10}, {10, 0}}
	small := Polygon{{2, 2}, {2, 3}, {3, 3}, {3, 2}}
	cross := Polygon{{5, 5}, {5, 15}, {15, 15}, {15, 5}}
	far := Polygon{{20, 20}, {20, 21}, {21, 21}, {21, 20}}
	if !big.ContainsPolygon(small) || big.ContainsPolygon(cross) || small.ContainsPolygon(big) {
		t.Fatal("ContainsPolygon")
	}
	if !big.Intersects(small) || !small.Intersects(big) || !big.Intersects(cross) || big.Intersects(far) {
		t.Fatal("Intersects")
	}
	if !big.Overlaps(cross) || big.Overlaps(small) || big.Overlaps(far) {
		t.Fatal("Overlaps")
	}
}

func TestPolygonWithHoles(t *testing.T) {
	p := PolygonWithHoles{
		Outer: Polygon{{0, 0}, {0, 10}, {10, 10}, {10, 0}},
		Holes: []Polygon{{{4, 4}, {4, 6}, {6, 6}, {6, 4}}},
	}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if !p.Contains(Coordinate{1, 1}) || p.Contains(Coordinate{5, 5}) || p.Contains(Coordinate{11, 5}) {
		t.Fatal("Contains")
	}
	near(t, "Area", float64(p.Area()), float64(p.Outer.Area()-p.Holes[0].Area()), 1e-3)
	if (PolygonWithHoles{Outer: Polygon{{0, 0}}}).Validate() == nil {
		t.Fatal("bad outer")
	}
	if (PolygonWithHoles{Outer: p.Outer, Holes: []Polygon{{{0, 0}}}}).Validate() == nil {
		t.Fatal("bad hole")
	}
	var _ Fence = p
}

func TestGeoJSONGeometries(t *testing.T) {
	poly := PolygonWithHoles{
		Outer: Polygon{{0, 0}, {0, 10}, {10, 10}, {10, 0}},
		Holes: []Polygon{{{4, 4}, {4, 6}, {6, 6}, {6, 4}}},
	}
	geoms := []Geometry{
		Coordinate{1, 2},
		MultiPoint{{1, 2}, {3, 4}},
		Polyline{{1, 2}, {3, 4}},
		MultiLineString{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}},
		poly,
		MultiPolygon{poly, {Outer: Polygon{{20, 20}, {20, 21}, {21, 21}}}},
		GeometryCollection{Coordinate{1, 2}, Polyline{{1, 2}, {3, 4}}},
	}
	for _, g := range geoms {
		b, err := MarshalGeometry(g)
		if err != nil {
			t.Fatalf("%T: %v", g, err)
		}
		back, err := ParseGeometry(b)
		if err != nil {
			t.Fatalf("%T parse %s: %v", g, b, err)
		}
		b2, _ := MarshalGeometry(back)
		if string(b) != string(b2) {
			t.Fatalf("%T round trip:\n%s\n%s", g, b, b2)
		}
	}
	b, _ := MarshalGeometry(poly)
	// exterior CCW, hole CW (RFC 7946); positions are [lng,lat]
	want := `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,10],[0,0]],[[4,4],[4,6],[6,6],[6,4],[4,4]]]}`
	if string(b) != want {
		t.Fatalf("\n%s\n%s", b, want)
	}
	for _, bad := range []string{
		`{"type":"Circle","coordinates":[]}`,
		`{"type":"Point","coordinates":[999,0]}`,
		`{"type":"LineString","coordinates":[[0,0]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,1],[0,0]]]}`,
		`{"type":"Polygon","coordinates":[]}`,
		`{"type":"MultiPoint","coordinates":"x"}`,
		`{"type":"GeometryCollection","geometries":[{"type":"Nope"}]}`,
		`[`,
	} {
		if _, err := ParseGeometry([]byte(bad)); !errors.Is(err, ErrParse) && !errors.Is(err, ErrInvalidLongitude) && !errors.Is(err, ErrInvalidPolygon) {
			t.Errorf("%s: %v", bad, err)
		}
	}
	if _, err := MarshalGeometry(nil); err == nil {
		t.Fatal("nil geometry")
	}
}

func TestFeatureCollection(t *testing.T) {
	fc := FeatureCollection{Features: []Feature{
		{ID: "shop-1", Geometry: tashkent, Properties: map[string]any{"name": "Chorsu"}},
		{Geometry: Polyline{tashkent, samarkand}},
	}}
	b, err := json.Marshal(fc)
	if err != nil {
		t.Fatal(err)
	}
	var back FeatureCollection
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err, string(b))
	}
	if len(back.Features) != 2 || back.Features[0].ID != "shop-1" || back.Features[0].Properties["name"] != "Chorsu" {
		t.Fatalf("%+v", back)
	}
	if c, ok := back.Features[0].Geometry.(Coordinate); !ok || c != tashkent {
		t.Fatal(back.Features[0].Geometry)
	}
	var f Feature
	if err := json.Unmarshal([]byte(`{"type":"Feature","geometry":null,"properties":null}`), &f); err != nil || f.Geometry != nil {
		t.Fatal("null geometry", err)
	}
	for _, bad := range []string{`{"type":"Nope"}`, `{"type":"Feature","geometry":{"type":"X"}}`} {
		if err := json.Unmarshal([]byte(bad), &f); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
	if err := json.Unmarshal([]byte(`{"type":"Feature"}`), &back); err == nil {
		t.Fatal("wrong collection type")
	}
}

func TestWKB(t *testing.T) {
	c := Coordinate{2, 1}
	const wkbLE = "0101000000000000000000F03F0000000000000040"
	const ewkb = "0101000020E6100000000000000000F03F0000000000000040"
	if got := c.EWKBHex(4326); got != ewkb {
		t.Fatal(got)
	}
	for _, in := range []any{wkbLE, ewkb, []byte(ewkb), mustHex(ewkb),
		"00000000013FF00000000000004000000000000000"} { // big-endian
		var got Coordinate
		if err := got.Scan(in); err != nil || got != c {
			t.Errorf("%v: %v %v", in, got, err)
		}
	}
	for _, bad := range []string{"0102000000", "01010000", "zz01000000000000000000F03F0000000000000040", "0101000000000000000000F03F00000000000059C0" + "00"} {
		var got Coordinate
		if err := got.Scan(bad); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
}

func mustHex(s string) []byte {
	b := make([]byte, len(s)/2)
	for i := range b {
		var v byte
		for _, ch := range s[2*i : 2*i+2] {
			v <<= 4
			switch {
			case ch >= '0' && ch <= '9':
				v |= byte(ch - '0')
			default:
				v |= byte(ch-'A') + 10
			}
		}
		b[i] = v
	}
	return b
}

func TestTrackSmooth(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var truth Polyline
	var tr Track
	for i := 0; i < 120; i++ {
		p := Destination(tashkent, Distance(i*10), 90) // 10 m/s east
		truth = append(truth, p)
		noisy := Destination(p, Distance(rng.NormFloat64()*15), Angle(rng.Float64()*360))
		tr = append(tr, TrackPoint{Coordinate: noisy, Time: t0.Add(time.Duration(i) * time.Second), Accuracy: 15})
	}
	s := tr.Smooth(1)
	if len(s) != len(tr) || s[0].Coordinate != tr[0].Coordinate {
		t.Fatal("shape")
	}
	errOf := func(x Track) (sum float64) {
		for i := 20; i < len(x); i++ { // skip warm-up
			sum += DistanceBetween(x[i].Coordinate, truth[i]).Meters()
		}
		return
	}
	if errOf(s) >= errOf(tr)*0.8 {
		t.Fatalf("smoothing did not reduce error: raw %.0f smooth %.0f", errOf(tr), errOf(s))
	}
	if (Track{}).Smooth(1) != nil {
		t.Fatal("empty")
	}
}

func TestGridIndex(t *testing.T) {
	idx := NewGridIndex[int](1 * Kilometer)
	rng := rand.New(rand.NewSource(2))
	pts := map[int]Coordinate{}
	for i := 0; i < 2000; i++ {
		c := Destination(tashkent, Distance(rng.Float64()*20000), Angle(rng.Float64()*360))
		pts[i] = c
		idx.Insert(i, c)
	}
	idx.Insert(0, pts[0]) // re-insert moves, does not duplicate
	if idx.Len() != 2000 {
		t.Fatal(idx.Len())
	}
	got := idx.WithinRadius(tashkent, 3*Kilometer)
	var want []int
	for id, c := range pts {
		if WithinRadius(tashkent, c, 3*Kilometer) {
			want = append(want, id)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("got %d want %d", len(got), len(want))
	}
	if !sort.SliceIsSorted(got, func(i, j int) bool {
		return DistanceBetween(tashkent, pts[got[i]]) < DistanceBetween(tashkent, pts[got[j]])
	}) {
		t.Fatal("not sorted by distance")
	}
	id, ok := idx.Nearest(tashkent, 5*Kilometer)
	if !ok || id != got[0] {
		t.Fatal("Nearest", id, ok)
	}
	idx.Remove(got[0])
	if idx.Len() != 1999 {
		t.Fatal("Remove")
	}
	if id2, _ := idx.Nearest(tashkent, 5*Kilometer); id2 == id {
		t.Fatal("removed still found")
	}
	idx.Remove(-1) // no-op
	if _, ok := NewGridIndex[int](1*Kilometer).Nearest(tashkent, 1); ok {
		t.Fatal("empty nearest")
	}
	// antimeridian and pole
	w := NewGridIndex[string](10 * Kilometer)
	w.Insert("east", Coordinate{0, 179.99})
	w.Insert("west", Coordinate{0, -179.99})
	w.Insert("pole", Coordinate{89.99, 45})
	if len(w.WithinRadius(Coordinate{0, 180}, 5*Kilometer)) != 2 {
		t.Fatal("antimeridian")
	}
	if r := w.WithinRadius(Coordinate{90, 0}, 5*Kilometer); len(r) != 1 || r[0] != "pole" {
		t.Fatal("pole", r)
	}
}

func FuzzParseGeometry(f *testing.F) {
	f.Add(`{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,0]]]}`)
	f.Add(`{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]}]}`)
	f.Fuzz(func(t *testing.T, s string) {
		g, err := ParseGeometry([]byte(s))
		if err == nil {
			if _, err := MarshalGeometry(g); err != nil {
				t.Fatalf("parsed but cannot marshal %q: %v", s, err)
			}
		}
	})
}

func FuzzScan(f *testing.F) {
	f.Add("0101000020E6100000000000000000F03F0000000000000040")
	f.Add("POINT(1 2)")
	f.Fuzz(func(t *testing.T, s string) {
		var c Coordinate
		if err := c.Scan(s); err == nil && !c.IsValid() {
			t.Fatalf("%q -> %v", s, c)
		}
		_ = c.Scan([]byte(s))
	})
}
