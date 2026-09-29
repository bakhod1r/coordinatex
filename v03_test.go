package coordinatex

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestUTM(t *testing.T) {
	u, err := ToUTM(Coordinate{0, 0})
	if err != nil || u.Zone != 31 || u.Band != 'N' || !u.North {
		t.Fatal(u, err)
	}
	near(t, "easting 0,0", u.Easting, 166021.443, 0.01)
	near(t, "northing 0,0", u.Northing, 0, 0.01)

	u, _ = ToUTM(london)
	if u.Zone != 30 || u.Band != 'U' {
		t.Fatal(u)
	}
	near(t, "london E", u.Easting, 699316, 2)
	near(t, "london N", u.Northing, 5710164, 2)
	if s := u.String(); s != "30U 699316 5710164" {
		t.Fatal(s)
	}

	s, _ := ToUTM(Coordinate{-33.8688, 151.2093})
	if s.North || s.Zone != 56 || s.Band != 'H' {
		t.Fatal(s)
	}
	for _, c := range []Coordinate{tashkent, samarkand, london, paris, {-33.8688, 151.2093}, {0, 0}, {83.9, 10}, {-79.9, -170}} {
		u, err := ToUTM(c)
		if err != nil {
			t.Fatal(err)
		}
		back, err := u.Coordinate()
		if err != nil {
			t.Fatal(err)
		}
		nearCoord(t, "UTM round trip "+c.String(), back, c, 1e-8)
	}
	// Norway / Svalbard exceptions
	if u, _ := ToUTM(Coordinate{60, 5}); u.Zone != 32 {
		t.Fatal("Norway zone", u.Zone)
	}
	if u, _ := ToUTM(Coordinate{78, 15}); u.Zone != 33 {
		t.Fatal("Svalbard zone", u.Zone)
	}
	if _, err := ToUTM(Coordinate{85, 0}); !errors.Is(err, ErrOutOfRange) {
		t.Fatal("polar", err)
	}
	if _, err := (UTM{Zone: 0}).Coordinate(); !errors.Is(err, ErrOutOfRange) {
		t.Fatal("bad zone", err)
	}
	p, err := ParseUTM("30U 699316 5710164")
	if err != nil || p.Zone != 30 || !p.North {
		t.Fatal(p, err)
	}
	for _, bad := range []string{"", "30 1 2", "99U 1 2", "30U x 2", "30I 1 2"} {
		if _, err := ParseUTM(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestMGRS(t *testing.T) {
	m, err := ToMGRS(Coordinate{0, 0}, 5)
	if err != nil || m != "31NAA6602100000" {
		t.Fatal(m, err)
	}
	for _, c := range []Coordinate{tashkent, samarkand, london, paris, {-33.8688, 151.2093}, {0, 0}, {60, 5}} {
		for _, prec := range []int{5, 3, 1} {
			m, err := ToMGRS(c, prec)
			if err != nil {
				t.Fatal(err)
			}
			back, err := ParseMGRS(m)
			if err != nil {
				t.Fatalf("%s: %v", m, err)
			}
			tol := map[int]Distance{5: 2, 3: 150, 1: 15000}[prec]
			if d := DistanceBetween(back, c); d > tol {
				t.Fatalf("%s: %v off by %v", m, c, d)
			}
		}
	}
	if m, _ := ToMGRS(london, 0); len(m) != len("30UXC") {
		t.Fatal("prec 0", m)
	}
	for _, bad := range []string{"", "31N", "31NAA123", "31NII0000", "99NAA00", "31NAAxx"} {
		if _, err := ParseMGRS(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestWebMercatorTiles(t *testing.T) {
	x, y := ToWebMercator(Coordinate{0, 0})
	near(t, "x0", x, 0, 1e-9)
	near(t, "y0", y, 0, 1e-6)
	x, _ = ToWebMercator(Coordinate{0, 180})
	near(t, "x180", x, 20037508.342789244, 1e-6)
	c := FromWebMercator(ToWebMercator(tashkent))
	nearCoord(t, "mercator round trip", c, tashkent, 1e-9)

	tile := TileAt(Coordinate{0, 0}, 1)
	if tile != (Tile{X: 1, Y: 1, Z: 1}) {
		t.Fatal(tile)
	}
	tile = TileAt(Coordinate{41.2995, 69.2401}, 10)
	if tile != (Tile{X: 708, Y: 382, Z: 10}) {
		t.Fatal(tile)
	}
	if !tile.Bounds().Contains(tashkent) {
		t.Fatal("tile bounds")
	}
	if q := (Tile{X: 3, Y: 5, Z: 3}).Quadkey(); q != "213" {
		t.Fatal(q)
	}
	back, err := TileFromQuadkey("213")
	if err != nil || back != (Tile{X: 3, Y: 5, Z: 3}) {
		t.Fatal(back, err)
	}
	if _, err := TileFromQuadkey("219"); err == nil {
		t.Fatal("bad quadkey")
	}
	if TileAt(Coordinate{89, 180}, 2) != (Tile{X: 3, Y: 0, Z: 2}) {
		t.Fatal("clamp", TileAt(Coordinate{89, 180}, 2))
	}
	if p := tile.Parent(); p != (Tile{X: 354, Y: 191, Z: 9}) {
		t.Fatal(p)
	}
	ch := (Tile{X: 1, Y: 1, Z: 1}).Children()
	if ch[0] != (Tile{X: 2, Y: 2, Z: 2}) || ch[3] != (Tile{X: 3, Y: 3, Z: 2}) {
		t.Fatal(ch)
	}
	tiles := TilesCovering(Bounds{40, 68, 42, 70}, 8)
	if len(tiles) == 0 {
		t.Fatal("cover")
	}
	for _, tl := range tiles {
		if !tl.Bounds().Intersects(Bounds{40, 68, 42, 70}) {
			t.Fatal("extra tile", tl)
		}
	}
	if s := tile.String(); s != "10/708/382" {
		t.Fatal(s)
	}
}

func TestPlusCode(t *testing.T) {
	if s := EncodePlusCode(Coordinate{47.365590, 8.524997}, 10); s != "8FVC9G8F+6X" {
		t.Fatal(s)
	}
	if s := EncodePlusCode(Coordinate{47.365590, 8.524997}, 8); s != "8FVC9G8F+" {
		t.Fatal(s)
	}
	if s := EncodePlusCode(Coordinate{47.365590, 8.524997}, 4); s != "8FVC0000+" {
		t.Fatal(s)
	}
	b, err := DecodePlusCode("8FVC9G8F+6X")
	if err != nil || !b.Contains(Coordinate{47.365590, 8.524997}) {
		t.Fatal(b, err)
	}
	for _, c := range []Coordinate{tashkent, samarkand, london, paris, {-33.8688, 151.2093}, {0, 0}, {90, 180}, {-90, -180}} {
		for _, n := range []int{2, 4, 6, 8, 10, 11, 12} {
			code := EncodePlusCode(c, n)
			b, err := DecodePlusCode(code)
			if err != nil {
				t.Fatalf("%s: %v", code, err)
			}
			if !b.Expand(0.01).Contains(c.Normalize()) && c.Lat != 90 && c.Lat != -90 {
				t.Fatalf("%v %d %s %+v", c, n, code, b)
			}
		}
	}
	if !IsValidPlusCode("8FVC9G8F+6X") || IsValidPlusCode("8FVC9G8F6X") || IsValidPlusCode("8FVC9G8F+6") || IsValidPlusCode("") {
		t.Fatal("IsValidPlusCode")
	}
	for _, bad := range []string{"", "8FVC9G8F", "AFVC9G8F+6X", "8FVC0000+6X", "9G8F+6X"} {
		if _, err := DecodePlusCode(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestGPX(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	tr := Track{{Coordinate: tashkent, Time: t0}, {Coordinate: samarkand, Time: t0.Add(time.Hour)}}
	var buf bytes.Buffer
	if err := WriteGPX(&buf, "trip", tr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `<trkpt lat="41.2995" lon="69.2401">`) || !strings.Contains(buf.String(), "<name>trip</name>") {
		t.Fatal(buf.String())
	}
	back, err := ReadGPX(&buf)
	if err != nil || len(back) != 2 || back[1].Coordinate != samarkand || !back[1].Time.Equal(t0.Add(time.Hour)) {
		t.Fatal(back, err)
	}
	wpt := `<gpx><wpt lat="1" lon="2"></wpt><rte><rtept lat="3" lon="4"/></rte></gpx>`
	back, err = ReadGPX(strings.NewReader(wpt))
	if err != nil || len(back) != 2 || back[0].Coordinate != (Coordinate{1, 2}) {
		t.Fatal(back, err)
	}
	if _, err := ReadGPX(strings.NewReader(`<gpx><wpt lat="91" lon="0"/></gpx>`)); err == nil {
		t.Fatal("invalid lat")
	}
	if _, err := ReadGPX(strings.NewReader(`<gpx`)); err == nil {
		t.Fatal("bad xml")
	}
}

func TestKML(t *testing.T) {
	var buf bytes.Buffer
	err := WriteKML(&buf, "places",
		Feature{Geometry: tashkent, Properties: map[string]any{"name": "Tashkent <city>"}},
		Feature{Geometry: Polyline{tashkent, samarkand}},
		Feature{Geometry: square},
	)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{"<Point><coordinates>69.2401,41.2995</coordinates></Point>", "<LineString>", "<Polygon>", "Tashkent &lt;city&gt;", "<name>places</name>"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in\n%s", want, s)
		}
	}
	if err := WriteKML(&buf, "x", Feature{Geometry: MultiPoint{tashkent}}); err == nil {
		t.Fatal("unsupported geometry")
	}
}
