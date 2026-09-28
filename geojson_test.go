package coordinatex

import (
	"errors"
	"testing"
)

func TestGeoJSONPoint(t *testing.T) {
	b, err := Coordinate{41.3111, 69.2797}.MarshalGeoJSON()
	if err != nil || string(b) != `{"type":"Point","coordinates":[69.2797,41.3111]}` {
		t.Fatal(string(b), err)
	}
	c, err := ParseGeoJSONPoint([]byte(`{"type":"Point","coordinates":[69.2797,41.3111]}`))
	if err != nil || c != (Coordinate{41.3111, 69.2797}) {
		t.Fatal(c, err)
	}
	for _, bad := range []string{`{"type":"LineString","coordinates":[1,2]}`, `{"type":"Point","coordinates":[1]}`, `{`, `{"type":"Point","coordinates":[200,0]}`} {
		if _, err := ParseGeoJSONPoint([]byte(bad)); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
	if _, err := (Coordinate{91, 0}).MarshalGeoJSON(); !errors.Is(err, ErrInvalidLatitude) {
		t.Fatal(err)
	}
}

func TestGeoJSONLineAndPolygon(t *testing.T) {
	b, err := Polyline{{0, 0}, {1, 2}}.MarshalGeoJSON()
	if err != nil || string(b) != `{"type":"LineString","coordinates":[[0,0],[2,1]]}` {
		t.Fatal(string(b), err)
	}
	// ring is closed and wound counter-clockwise per RFC 7946
	const ccw = `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,1],[0,0]]]}`
	b, err = square.MarshalGeoJSON()
	if err != nil || string(b) != ccw {
		t.Fatal(string(b), err)
	}
	b, err = Polygon{{0, 0}, {1, 0}, {1, 1}, {0, 1}}.MarshalGeoJSON() // clockwise input is rewound
	if err != nil || string(b) != ccw {
		t.Fatal(string(b), err)
	}
	if _, err := (Polygon{{0, 0}}).MarshalGeoJSON(); err == nil {
		t.Fatal("invalid polygon")
	}
	if _, err := (Polyline{{100, 0}}).MarshalGeoJSON(); err == nil {
		t.Fatal("invalid line")
	}
}
