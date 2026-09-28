package coordinatex

import (
	"errors"
	"testing"
)

func TestNewBounds(t *testing.T) {
	b, err := NewBounds(Coordinate{40, 68}, Coordinate{42, 70})
	if err != nil {
		t.Fatal(err)
	}
	if !b.Contains(tashkent) || b.Contains(london) {
		t.Fatal("Contains")
	}
	nearCoord(t, "Center", b.Center(), Coordinate{41, 69}, 1e-9)
	near(t, "Height", b.Height().Kilometers(), 222.4, 0.5)
	if b.Width() <= 0 {
		t.Fatal("Width")
	}
	if _, err := NewBounds(Coordinate{42, 68}, Coordinate{40, 70}); !errors.Is(err, ErrInvalidBounds) {
		t.Fatalf("inverted lat: %v", err)
	}
	if _, err := NewBounds(Coordinate{100, 0}, Coordinate{0, 0}); !errors.Is(err, ErrInvalidLatitude) {
		t.Fatalf("invalid sw: %v", err)
	}
	if _, err := NewBounds(Coordinate{0, 0}, Coordinate{0, 200}); !errors.Is(err, ErrInvalidLongitude) {
		t.Fatalf("invalid ne: %v", err)
	}
}

func TestBoundsAntimeridian(t *testing.T) {
	b, err := NewBounds(Coordinate{-10, 170}, Coordinate{10, -170})
	if err != nil {
		t.Fatal(err)
	}
	if !b.CrossesAntimeridian() {
		t.Fatal("expected crossing")
	}
	if !b.Contains(Coordinate{0, 179}) || !b.Contains(Coordinate{0, -175}) || b.Contains(Coordinate{0, 0}) {
		t.Fatal("Contains across antimeridian")
	}
	nearCoord(t, "Center", b.Center(), Coordinate{0, -180}, 1e-9)
	near(t, "Width", b.Width().Kilometers(), 2223.9, 1)
}

func TestBoundsOf(t *testing.T) {
	if _, err := BoundsOf(); !errors.Is(err, ErrEmpty) {
		t.Fatal(err)
	}
	b, err := BoundsOf(tashkent, samarkand, london)
	if err != nil {
		t.Fatal(err)
	}
	want := Bounds{MinLat: 39.6542, MinLng: -0.1278, MaxLat: 51.5074, MaxLng: 69.2401}
	if b != want {
		t.Fatalf("%+v", b)
	}
	if _, err := BoundsOf(Coordinate{91, 0}); err == nil {
		t.Fatal("invalid point accepted")
	}
}

func TestBoundsAround(t *testing.T) {
	b := BoundsAround(tashkent, 5*Kilometer)
	for _, brg := range []Angle{0, 45, 90, 180, 270} {
		p := Destination(tashkent, 4999*Meter, brg)
		if !b.Contains(p) {
			t.Errorf("bearing %v: %v not in %+v", brg, p, b)
		}
	}
	if b.Contains(Destination(tashkent, 6*Kilometer, 0)) {
		t.Error("too big")
	}
	// pole: lng spans full
	p := BoundsAround(Coordinate{89.99, 0}, 10*Kilometer)
	if p.MaxLat != 90 || p.MinLng != -180 || p.MaxLng != 180 {
		t.Fatalf("pole: %+v", p)
	}
	// antimeridian
	a := BoundsAround(Coordinate{0, 179.99}, 10*Kilometer)
	if !a.CrossesAntimeridian() || !a.Contains(Coordinate{0, -179.99}) {
		t.Fatalf("antimeridian: %+v", a)
	}
}

func TestBoundsIntersectsUnionExpand(t *testing.T) {
	a := Bounds{0, 0, 10, 10}
	b := Bounds{5, 5, 15, 15}
	c := Bounds{20, 20, 30, 30}
	if !a.Intersects(b) || a.Intersects(c) {
		t.Fatal("Intersects")
	}
	x := Bounds{-5, 170, 5, -170}
	if !x.Intersects(Bounds{0, 175, 1, 176}) || !x.Intersects(Bounds{0, -175, 1, -174}) || x.Intersects(a) {
		t.Fatal("Intersects antimeridian")
	}
	if u := a.Union(c); u != (Bounds{0, 0, 30, 30}) {
		t.Fatalf("Union %+v", u)
	}
	e := a.Expand(1 * Kilometer)
	if e.MinLat >= a.MinLat || e.MaxLat <= a.MaxLat || e.MinLng >= a.MinLng || e.MaxLng <= a.MaxLng {
		t.Fatalf("Expand %+v", e)
	}
	nearCoord(t, "SW", a.SouthWest(), Coordinate{0, 0}, 0)
	nearCoord(t, "NE", a.NorthEast(), Coordinate{10, 10}, 0)
}
