package coordinatex

import (
	"errors"
	"reflect"
	"testing"
)

func TestBoundsFilter(t *testing.T) {
	b := Bounds{MinLat: 40, MinLng: 68, MaxLat: 42, MaxLng: 70}
	f, err := Postgres.BoundsFilter("lat", "lng", b, 0)
	if err != nil {
		t.Fatal(err)
	}
	if f.Clause != "(lat BETWEEN $1 AND $2 AND lng BETWEEN $3 AND $4)" || !reflect.DeepEqual(f.Args, []any{40.0, 42.0, 68.0, 70.0}) {
		t.Fatalf("%+v", f)
	}
	f, _ = MySQL.BoundsFilter("s.lat", "s.lng", b, 0)
	if f.Clause != "(s.lat BETWEEN ? AND ? AND s.lng BETWEEN ? AND ?)" {
		t.Fatal(f.Clause)
	}
	x := Bounds{MinLat: -10, MinLng: 170, MaxLat: 10, MaxLng: -170}
	f, _ = Postgres.BoundsFilter("lat", "lng", x, 2)
	if f.Clause != "(lat BETWEEN $3 AND $4 AND (lng >= $5 OR lng <= $6))" || !reflect.DeepEqual(f.Args, []any{-10.0, 10.0, 170.0, -170.0}) {
		t.Fatalf("%+v", f)
	}
	for _, col := range []string{"", "lat;DROP TABLE x", "1lat", "a.b.c", "lat lng", `"lat"`} {
		if _, err := Postgres.BoundsFilter(col, "lng", b, 0); !errors.Is(err, ErrInvalidIdentifier) {
			t.Errorf("%q: %v", col, err)
		}
	}
	if _, err := Postgres.BoundsFilter("lat", "lng", b, -1); err == nil {
		t.Fatal("negative offset")
	}
}

func TestRadiusFilter(t *testing.T) {
	f, err := Postgres.RadiusFilter("lat", "lng", tashkent, 5*Kilometer, 0)
	if err != nil {
		t.Fatal(err)
	}
	b := BoundsAround(tashkent, 5*Kilometer)
	want, _ := Postgres.BoundsFilter("lat", "lng", b, 0)
	if !reflect.DeepEqual(f, want) {
		t.Fatalf("%+v", f)
	}
}

func TestPostGISFilters(t *testing.T) {
	f, err := Postgres.DWithin("geog", tashkent, 5*Kilometer, 1)
	if err != nil {
		t.Fatal(err)
	}
	if f.Clause != "ST_DWithin(geog, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4)" ||
		!reflect.DeepEqual(f.Args, []any{tashkent.Lng, tashkent.Lat, 5000.0}) {
		t.Fatalf("%+v", f)
	}
	f, _ = Postgres.DistanceOrder("geog", tashkent, 0)
	if f.Clause != "geog <-> ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography" {
		t.Fatal(f.Clause)
	}
	f, _ = Postgres.Envelope("geom", Bounds{40, 68, 42, 70}, 0)
	if f.Clause != "geom && ST_MakeEnvelope($1, $2, $3, $4, 4326)" || !reflect.DeepEqual(f.Args, []any{68.0, 40.0, 70.0, 42.0}) {
		t.Fatalf("%+v", f)
	}
	f, _ = Postgres.Envelope("geom", Bounds{-10, 170, 10, -170}, 0)
	if f.Clause != "(geom && ST_MakeEnvelope($1, $2, $3, $4, 4326) OR geom && ST_MakeEnvelope($5, $6, $7, $8, 4326))" ||
		!reflect.DeepEqual(f.Args, []any{170.0, -10.0, 180.0, 10.0, -180.0, -10.0, -170.0, 10.0}) {
		t.Fatalf("%+v", f)
	}
	if _, err := Postgres.DWithin("geog; --", tashkent, 1, 0); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatal(err)
	}
	if _, err := Postgres.DWithin("geog", Coordinate{91, 0}, 1, 0); !errors.Is(err, ErrInvalidLatitude) {
		t.Fatal(err)
	}
	if _, err := Postgres.DistanceOrder("x y", tashkent, 0); err == nil {
		t.Fatal("bad column")
	}
	if _, err := Postgres.Envelope("x y", Bounds{}, 0); err == nil {
		t.Fatal("bad column")
	}
}

func TestGeohashFilter(t *testing.T) {
	f, err := MySQL.GeohashPrefixFilter("gh", []string{"tzq2b", "tzq2c"}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if f.Clause != "(gh LIKE ? OR gh LIKE ?)" || !reflect.DeepEqual(f.Args, []any{"tzq2b%", "tzq2c%"}) {
		t.Fatalf("%+v", f)
	}
	if _, err := MySQL.GeohashPrefixFilter("gh", []string{"a%"}, 0); !errors.Is(err, ErrInvalidGeohash) {
		t.Fatal("invalid cell", err)
	}
	if _, err := MySQL.GeohashPrefixFilter("gh", nil, 0); !errors.Is(err, ErrEmpty) {
		t.Fatal(err)
	}
	if _, err := MySQL.GeohashPrefixFilter("g h", []string{"u"}, 0); !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatal(err)
	}
}

func TestSQLFilterAnd(t *testing.T) {
	a, _ := Postgres.DWithin("geog", tashkent, 100, 0)
	b, _ := Postgres.GeohashPrefixFilter("gh", []string{"tzq"}, len(a.Args))
	c := a.And(b)
	if c.Clause != a.Clause+" AND "+b.Clause || len(c.Args) != 4 || c.Args[3] != "tzq%" {
		t.Fatalf("%+v", c)
	}
	if b.Clause != "(gh LIKE $4)" {
		t.Fatal(b.Clause)
	}
}
