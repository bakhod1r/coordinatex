package coordinatex

import (
	"errors"
	"testing"
)

var square = Polygon{{0, 0}, {0, 1}, {1, 1}, {1, 0}}

func TestPolygonContains(t *testing.T) {
	if !square.Contains(Coordinate{0.5, 0.5}) || square.Contains(Coordinate{1.5, 0.5}) || square.Contains(Coordinate{0.5, -0.1}) {
		t.Fatal("square")
	}
	closed := append(Polygon{}, square...)
	closed = append(closed, square[0])
	if !closed.Contains(Coordinate{0.5, 0.5}) {
		t.Fatal("closed ring")
	}
	// concave "C" shape
	c := Polygon{{0, 0}, {0, 3}, {1, 3}, {1, 1}, {2, 1}, {2, 3}, {3, 3}, {3, 0}}
	if c.Contains(Coordinate{1.5, 2}) || !c.Contains(Coordinate{0.5, 2}) {
		t.Fatal("concave")
	}
	if (Polygon{{0, 0}, {1, 1}}).Contains(Coordinate{0.5, 0.5}) {
		t.Fatal("degenerate")
	}
}

func TestPolygonMeasures(t *testing.T) {
	near(t, "Area km2", square.Area().SquareKilometers(), 12364, 5)
	near(t, "Area orientation-free", Polygon{{0, 0}, {1, 0}, {1, 1}, {0, 1}}.Area().SquareMeters(), float64(square.Area()), 1e-3)
	near(t, "Perimeter km", square.Perimeter().Kilometers(), 4*111.19, 0.2)
	nearCoord(t, "Centroid", square.Centroid(), Coordinate{0.5, 0.5}, 1e-9)
	if (Polygon{}).Area() != 0 || (Polygon{}).Perimeter() != 0 {
		t.Fatal("empty")
	}
	nearCoord(t, "Centroid degenerate", Polygon{{0, 0}, {0, 2}, {0, 4}}.Centroid(), Coordinate{0, 2}, 1e-9)
	nearCoord(t, "Centroid empty", Polygon{}.Centroid(), Coordinate{}, 0)
}

func TestPolygonOrientationBounds(t *testing.T) {
	// (lat,lng) 0,0 -> 0,1 -> 1,1 -> 1,0 goes east, north, west: counter-clockwise on a map
	if square.IsClockwise() {
		t.Fatal("expected counter-clockwise")
	}
	if !(Polygon{{0, 0}, {1, 0}, {1, 1}, {0, 1}}).IsClockwise() {
		t.Fatal("expected clockwise")
	}
	b, err := square.Bounds()
	if err != nil || b != (Bounds{0, 0, 1, 1}) {
		t.Fatal(b, err)
	}
	if _, err := (Polygon{}).Bounds(); !errors.Is(err, ErrEmpty) {
		t.Fatal(err)
	}
	if err := square.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Polygon{{0, 0}, {1, 1}}).Validate(); !errors.Is(err, ErrInvalidPolygon) {
		t.Fatal(err)
	}
	if err := (Polygon{{0, 0}, {1, 1}, {100, 0}}).Validate(); !errors.Is(err, ErrInvalidLatitude) {
		t.Fatal(err)
	}
}
