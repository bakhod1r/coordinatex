package coordinatex

import "testing"

func TestFences(t *testing.T) {
	circle := CircleFence{Center: tashkent, Radius: 10 * Kilometer}
	rect := RectangleFence{Bounds: Bounds{40, 68, 42, 70}}
	poly := PolygonFence{Polygon: square}

	if !circle.Contains(tashkent) || circle.Contains(samarkand) {
		t.Fatal("circle")
	}
	if !rect.Contains(tashkent) || rect.Contains(london) {
		t.Fatal("rect")
	}
	if !poly.Contains(Coordinate{0.5, 0.5}) || poly.Contains(tashkent) {
		t.Fatal("poly")
	}
	anyF := AnyFence(circle, poly)
	if !anyF.Contains(tashkent) || !anyF.Contains(Coordinate{0.5, 0.5}) || anyF.Contains(london) {
		t.Fatal("any")
	}
	allF := AllFence(circle, rect)
	if !allF.Contains(tashkent) || allF.Contains(Coordinate{40.5, 68.5}) {
		t.Fatal("all")
	}
	if AnyFence().Contains(tashkent) || !AllFence().Contains(tashkent) {
		t.Fatal("empty composition")
	}
	var f Fence = FenceFunc(func(c Coordinate) bool { return c.Lat > 0 })
	if !f.Contains(tashkent) {
		t.Fatal("FenceFunc")
	}
}
