package coordinatex

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

var (
	_ driver.Valuer = Coordinate{}
	_ sql.Scanner   = (*Coordinate)(nil)
)

func TestWKT(t *testing.T) {
	c := Coordinate{41.3111, 69.2797}
	if s := c.WKT(); s != "POINT(69.2797 41.3111)" {
		t.Fatal(s)
	}
	v, err := c.Value()
	if err != nil || v != "POINT(69.2797 41.3111)" {
		t.Fatal(v, err)
	}
	if _, err := (Coordinate{0, 200}).Value(); err == nil {
		t.Fatal("invalid valued")
	}
	for _, in := range []any{"POINT(69.2797 41.3111)", []byte("point ( 69.2797 41.3111 )"), "SRID=4326;POINT(69.2797 41.3111)"} {
		var got Coordinate
		if err := got.Scan(in); err != nil || got != c {
			t.Errorf("%v: %v %v", in, got, err)
		}
	}
	for _, bad := range []any{nil, 42, "LINESTRING(0 0,1 1)", "POINT(1)", "POINT(a b)", "POINT(0 100)"} {
		var got Coordinate
		if err := got.Scan(bad); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
	if _, err := ParseWKT("POINT(1 2"); !errors.Is(err, ErrParse) {
		t.Fatal(err)
	}
}
