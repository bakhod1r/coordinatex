package coordinatex

import "testing"

func TestWithinRadius(t *testing.T) {
	if !WithinRadius(tashkent, Destination(tashkent, 4*Kilometer, 30), 5*Kilometer) {
		t.Fatal("inside")
	}
	if WithinRadius(tashkent, samarkand, 5*Kilometer) {
		t.Fatal("outside")
	}
}

func TestNearest(t *testing.T) {
	pts := []Coordinate{london, samarkand, paris}
	c, i, ok := Nearest(tashkent, pts)
	if !ok || i != 1 || c != samarkand {
		t.Fatal(c, i, ok)
	}
	if _, _, ok := Nearest(tashkent, nil); ok {
		t.Fatal("empty")
	}
}

func TestNearestNAndSort(t *testing.T) {
	pts := []Coordinate{london, samarkand, paris}
	got := NearestN(tashkent, pts, 2)
	if len(got) != 2 || got[0] != samarkand || got[1] != paris {
		t.Fatal(got)
	}
	if pts[0] != london {
		t.Fatal("NearestN mutated input")
	}
	if len(NearestN(tashkent, pts, 10)) != 3 || NearestN(tashkent, pts, 0) != nil {
		t.Fatal("n bounds")
	}
	SortByDistance(paris, pts)
	if pts[0] != paris || pts[1] != london || pts[2] != samarkand {
		t.Fatal(pts)
	}
}

func TestWithinRadiusFilter(t *testing.T) {
	pts := []Coordinate{london, samarkand, paris}
	got := WithinRadiusFilter(paris, pts, 400*Kilometer)
	if len(got) != 2 || got[0] != london || got[1] != paris {
		t.Fatal(got)
	}
}
