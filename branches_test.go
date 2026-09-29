package coordinatex

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

// Edge and error branches not exercised by the feature tests.

func TestBoundsBranches(t *testing.T) {
	wide := Bounds{0, -175, 1, 175}.Expand(1000 * Kilometer)
	if wide.MinLng != -180 || wide.MaxLng != 180 {
		t.Fatal("expand to full longitude", wide)
	}
	if _, ok := (Bounds{0, 0, 10, 10}).Intersection(Bounds{0, 20, 10, 30}); ok {
		t.Fatal("lng-disjoint")
	}
	if len(IntermediatePoints(tashkent, samarkand, -1)) != 2 {
		t.Fatal("negative n")
	}
}

func TestVincentyBranches(t *testing.T) {
	d, err := DistanceVincenty(Coordinate{0, 0}, Coordinate{0, 10})
	if err != nil {
		t.Fatal(err)
	}
	near(t, "equatorial", d.Kilometers(), 1113.19, 0.01)
	if _, err := DestinationVincenty(Coordinate{91, 0}, 1, 0); !errors.Is(err, ErrInvalidLatitude) {
		t.Fatal(err)
	}
}

func TestParseBranches(t *testing.T) {
	for _, in := range []string{
		"41.3,abc",                            // longitude fails
		"41N,69N",                             // wrong axis letter for longitude
		"1" + strings.Repeat("0", 400) + ",0", // float out of range
		"41 -18N 69E",                         // negative minutes
		"N,69",                                // no number
		"1 2 3 4,5",                           // too many parts
		"-41N,69",                             // sign and hemisphere
	} {
		if _, err := Parse(in); err == nil {
			t.Errorf("%q accepted", in)
		}
	}
}

func TestTransitionStrings(t *testing.T) {
	for tr, want := range map[Transition]string{StayedOutside: "stayed_outside", Entered: "entered", Exited: "exited", StayedInside: "stayed_inside"} {
		if tr.String() != want {
			t.Errorf("%d: %s", tr, tr.String())
		}
	}
}

func TestGeohashPoleNeighbors(t *testing.T) {
	h := EncodeGeohash(Coordinate{89.99, 0}, 3)
	n, err := GeohashNeighbors(h)
	if err != nil || n[0] != h {
		t.Fatal("north neighbour at pole is clamped to itself", n, err)
	}
}

func TestGeoJSONBranches(t *testing.T) {
	bad := Coordinate{91, 0}
	for _, g := range []Geometry{
		MultiPoint{bad},
		MultiLineString{{tashkent, bad}},
		MultiPolygon{{Outer: Polygon{{0, 0}}}},
		GeometryCollection{bad},
		PolygonWithHoles{Outer: square, Holes: []Polygon{{{0, 0}}}},
	} {
		if _, err := MarshalGeometry(g); err == nil {
			t.Errorf("%T accepted", g)
		}
	}
	for _, in := range []string{
		`{"type":"Point","coordinates":"x"}`,
		`{"type":"MultiPoint","coordinates":[[1,2],[1,95]]}`,
		`{"type":"LineString","coordinates":"x"}`,
		`{"type":"MultiLineString","coordinates":"x"}`,
		`{"type":"MultiLineString","coordinates":[[[0,0]]]}`,
		`{"type":"Polygon","coordinates":"x"}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,1],[0,95],[0,0]]]}`,
		`{"type":"MultiPolygon","coordinates":"x"}`,
		`{"type":"MultiPolygon","coordinates":[[]]}`,
	} {
		if _, err := ParseGeometry([]byte(in)); err == nil {
			t.Errorf("%s accepted", in)
		}
	}
	if _, err := ParseGeoJSONPoint([]byte(`{"type":"LineString","coordinates":[[0,0],[1,1]]}`)); !errors.Is(err, ErrParse) {
		t.Fatal(err)
	}
	if _, err := json.Marshal(Feature{Geometry: bad}); err == nil {
		t.Fatal("invalid feature geometry")
	}
	var f Feature
	if err := f.UnmarshalJSON([]byte(`[`)); err == nil {
		t.Fatal("bad feature json")
	}
	b, _ := json.Marshal(FeatureCollection{})
	if string(b) != `{"type":"FeatureCollection","features":[]}` {
		t.Fatal(string(b))
	}
	var fc FeatureCollection
	if err := fc.UnmarshalJSON([]byte(`[`)); err == nil {
		t.Fatal("bad collection json")
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestGPXKMLBranches(t *testing.T) {
	if err := WriteGPX(&strings.Builder{}, "x", Track{{Coordinate: Coordinate{91, 0}}}); err == nil {
		t.Fatal("invalid point")
	}
	if err := WriteGPX(failWriter{}, "x", nil); err == nil {
		t.Fatal("writer error")
	}
	for _, doc := range []string{
		`<gpx><rte><rtept lat="91" lon="0"/></rte></gpx>`,
		`<gpx><trk><trkseg><trkpt lat="91" lon="0"/></trkseg></trk></gpx>`,
	} {
		if _, err := ReadGPX(strings.NewReader(doc)); err == nil {
			t.Errorf("%s accepted", doc)
		}
	}
	var sb strings.Builder
	holed := PolygonWithHoles{Outer: square, Holes: []Polygon{{{0.2, 0.2}, {0.2, 0.4}, {0.4, 0.4}}}}
	if err := WriteKML(&sb, "x", Feature{Geometry: holed}); err != nil || !strings.Contains(sb.String(), "<innerBoundaryIs>") {
		t.Fatal(sb.String(), err)
	}
}

func TestGridIndexSharedCell(t *testing.T) {
	g := NewGridIndex[int](10 * Kilometer)
	g.Insert(1, tashkent)
	g.Insert(2, Destination(tashkent, 1, 0))
	g.Remove(1)
	g.Insert(3, london)
	g.Remove(3) // last item in its cell: the cell is dropped
	if len(g.cells) != 1 {
		t.Fatal("empty cell kept")
	}
	if got := g.WithinRadius(tashkent, 10); len(got) != 1 || got[0] != 2 {
		t.Fatal(got)
	}
}

func TestPlusCodeBranches(t *testing.T) {
	c := Coordinate{47.365590, 8.524997}
	if EncodePlusCode(c, 1) != "8F000000+" || EncodePlusCode(c, 3) != "8FVC0000+" || len(EncodePlusCode(c, 20)) != 16 {
		t.Fatal(EncodePlusCode(c, 1), EncodePlusCode(c, 3), EncodePlusCode(c, 20))
	}
	if IsValidPlusCode("X2222222+") || IsValidPlusCode("2X222222+") {
		t.Fatal("out-of-range first digits")
	}
}

func TestGeometryBranches(t *testing.T) {
	if !segmentsIntersect(Coordinate{0, 0}, Coordinate{0, 2}, Coordinate{0, 1}, Coordinate{0, 3}) {
		t.Fatal("collinear overlap")
	}
	if segmentsIntersect(Coordinate{0, 0}, Coordinate{0, 1}, Coordinate{0, 2}, Coordinate{0, 3}) {
		t.Fatal("collinear disjoint")
	}
	if (Polygon{{0, 0}}).Intersects(square) {
		t.Fatal("degenerate intersects")
	}
	b, err := PolygonWithHoles{Outer: square}.Bounds()
	if err != nil || b != (Bounds{0, 0, 1, 1}) {
		t.Fatal(b, err)
	}
	if !math.IsInf(float64(Polygon{}.DistanceTo(tashkent)), 1) {
		t.Fatal("empty polygon distance")
	}
	p, _, _ := Polyline{{0, 0}, {0, 0}, {0, 1}}.NearestPoint(Coordinate{0, -1})
	if p != (Coordinate{0, 0}) {
		t.Fatal(p)
	}
	if Polyline(nil).SegmentLengths() != nil || Polyline(nil).Bearings() != nil {
		t.Fatal("short line")
	}
}

func TestPolylineDecodeBranches(t *testing.T) {
	if _, err := DecodePolyline("_p~iF~", 5); !errors.Is(err, ErrInvalidPolyline) {
		t.Fatal("truncated longitude", err)
	}
	if _, err := DecodePolyline(EncodePolyline(Polyline{{100, 0}}, 5), 5); !errors.Is(err, ErrInvalidPolyline) {
		t.Fatal("out of range", err)
	}
}

func TestDBSCANBorderPoint(t *testing.T) {
	// The first point has too few neighbours to be core, so it is marked Noise,
	// then absorbed as a border point when its cluster is expanded.
	core := Destination(tashkent, 150, 90)
	pts := []Coordinate{tashkent}
	for i := 0; i < 5; i++ {
		pts = append(pts, Destination(core, Distance(i*10), 0))
	}
	labels, n := DBSCAN(pts, 152*Meter, 5)
	if n != 1 || labels[0] != 0 {
		t.Fatal(labels, n)
	}
}

func TestTrackBranches(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if (Track{}).Speeds() != nil || (Track{}).FilterSpeed(1) != nil {
		t.Fatal("empty")
	}
	same := Track{{Coordinate: tashkent, Time: t0}, {Coordinate: samarkand, Time: t0}}
	if same.Speeds()[0] != 0 {
		t.Fatal("zero dt speed")
	}
	back := Track{{Coordinate: tashkent, Time: t0}, {Coordinate: tashkent, Time: t0.Add(-time.Second)}}
	if s := back.Smooth(1); len(s) != 2 || s[1].Accuracy <= 0 {
		t.Fatal("backwards time / default accuracy", s)
	}
	moving := Track{
		{Coordinate: tashkent, Time: t0},
		{Coordinate: tashkent, Time: t0.Add(10 * time.Minute)},
		{Coordinate: samarkand, Time: t0.Add(time.Hour)},
	}
	trips := moving.Trips(50, 5*time.Minute)
	if len(trips) != 1 || trips[0][0].Coordinate != samarkand {
		t.Fatal(trips)
	}
}

func TestSQLBranches(t *testing.T) {
	var c Coordinate
	if err := c.Scan([]byte{1, 2}); err == nil {
		t.Fatal("short binary WKB")
	}
	if err := c.Scan("0101000020E6"); err == nil {
		t.Fatal("truncated SRID")
	}
	if _, err := ParseWKT("POINT 1 2)"); !errors.Is(err, ErrParse) {
		t.Fatal(err)
	}
	if _, err := Postgres.DWithin("g", tashkent, 1, -1); err == nil {
		t.Fatal("DWithin offset")
	}
	if _, err := Postgres.DistanceOrder("g", tashkent, -1); err == nil {
		t.Fatal("DistanceOrder offset")
	}
	if _, err := Postgres.Envelope("g", Bounds{}, -1); err == nil {
		t.Fatal("Envelope offset")
	}
	if _, err := Postgres.GeohashPrefixFilter("g", []string{"u"}, -1); err == nil {
		t.Fatal("GeohashPrefixFilter offset")
	}
}

func TestProjectionBranches(t *testing.T) {
	if _, err := ToUTM(Coordinate{0, 200}); !errors.Is(err, ErrInvalidLongitude) {
		t.Fatal(err)
	}
	if _, err := ToMGRS(Coordinate{85, 0}, 5); !errors.Is(err, ErrOutOfRange) {
		t.Fatal(err)
	}
	if tl, err := TileFromQuadkey("0"); err != nil || tl != (Tile{Z: 1}) {
		t.Fatal(tl, err)
	}
}
