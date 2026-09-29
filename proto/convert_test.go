package coordinatexproto_test

import (
	"errors"
	"testing"

	"github.com/bakhod1r/coordinatex"
	cxp "github.com/bakhod1r/coordinatex/proto"
	"github.com/bakhod1r/coordinatex/proto/coordinatexpb"
	"google.golang.org/protobuf/proto"
)

var (
	tashkent  = coordinatex.MustNew(41.2995, 69.2401)
	samarkand = coordinatex.MustNew(39.6542, 66.9597)
)

func TestLatLngRoundTrip(t *testing.T) {
	m := cxp.FromCoordinate(tashkent)
	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var back coordinatexpb.LatLng
	if err := proto.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	c, err := cxp.ToCoordinate(&back)
	if err != nil || c != tashkent {
		t.Fatal(c, err)
	}
	if _, err := cxp.ToCoordinate(&coordinatexpb.LatLng{Lat: 91}); !errors.Is(err, coordinatex.ErrInvalidLatitude) {
		t.Fatal(err)
	}
	if _, err := cxp.ToCoordinate(nil); !errors.Is(err, cxp.ErrNilMessage) {
		t.Fatal(err)
	}
}

func TestBoundsCircle(t *testing.T) {
	b := coordinatex.Bounds{MinLat: -10, MinLng: 170, MaxLat: 10, MaxLng: -170}
	got, err := cxp.ToBounds(cxp.FromBounds(b))
	if err != nil || got != b {
		t.Fatal(got, err)
	}
	if _, err := cxp.ToBounds(&coordinatexpb.Bounds{MinLat: 10, MaxLat: 0}); !errors.Is(err, coordinatex.ErrInvalidBounds) {
		t.Fatal(err)
	}
	if _, err := cxp.ToBounds(nil); !errors.Is(err, cxp.ErrNilMessage) {
		t.Fatal(err)
	}
	c := coordinatex.Circle{Center: tashkent, Radius: 5 * coordinatex.Kilometer}
	gc, err := cxp.ToCircle(cxp.FromCircle(c))
	if err != nil || gc != c {
		t.Fatal(gc, err)
	}
	if _, err := cxp.ToCircle(&coordinatexpb.Circle{Center: cxp.FromCoordinate(tashkent), RadiusMeters: -1}); err == nil {
		t.Fatal("negative radius")
	}
	if _, err := cxp.ToCircle(&coordinatexpb.Circle{}); !errors.Is(err, cxp.ErrNilMessage) {
		t.Fatal(err)
	}
}

func TestPolylinePolygon(t *testing.T) {
	l := coordinatex.Polyline{tashkent, samarkand}
	gl, err := cxp.ToPolyline(cxp.FromPolyline(l))
	if err != nil || len(gl) != 2 || gl[1] != samarkand {
		t.Fatal(gl, err)
	}
	if _, err := cxp.ToPolyline(&coordinatexpb.Polyline{Points: []*coordinatexpb.LatLng{{Lat: 100}}}); err == nil {
		t.Fatal("invalid point")
	}
	p := coordinatex.PolygonWithHoles{
		Outer: coordinatex.Polygon{ll(0, 0), ll(0, 10), ll(10, 10), ll(10, 0)},
		Holes: []coordinatex.Polygon{{ll(4, 4), ll(4, 6), ll(6, 6)}},
	}
	gp, err := cxp.ToPolygon(cxp.FromPolygon(p))
	if err != nil || len(gp.Outer) != 4 || len(gp.Holes) != 1 || gp.Holes[0][2] != (coordinatex.Coordinate{Lat: 6, Lng: 6}) {
		t.Fatal(gp, err)
	}
	if _, err := cxp.ToPolygon(&coordinatexpb.Polygon{Outer: &coordinatexpb.Ring{}}); !errors.Is(err, coordinatex.ErrInvalidPolygon) {
		t.Fatal(err)
	}
	if _, err := cxp.ToPolygon(nil); !errors.Is(err, cxp.ErrNilMessage) {
		t.Fatal(err)
	}
	if _, err := cxp.ToPolyline(nil); !errors.Is(err, cxp.ErrNilMessage) {
		t.Fatal(err)
	}
}

func ll(lat, lng float64) coordinatex.Coordinate { return coordinatex.Coordinate{Lat: lat, Lng: lng} }
