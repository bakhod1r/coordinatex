package coordinatex

import "math"

// Circle is a spherical cap: every point within Radius of Center.
type Circle struct {
	Center Coordinate
	Radius Distance
}

// Contains reports whether c is inside or on the circle.
func (ci Circle) Contains(c Coordinate) bool { return WithinRadius(ci.Center, c, ci.Radius) }

// DistanceTo returns the distance from c to the circle edge; negative inside.
func (ci Circle) DistanceTo(c Coordinate) Distance { return DistanceBetween(ci.Center, c) - ci.Radius }

// Intersects reports whether two circles overlap or touch.
func (ci Circle) Intersects(o Circle) bool {
	return DistanceBetween(ci.Center, o.Center) <= ci.Radius+o.Radius
}

// Bounds returns a box enclosing the circle.
func (ci Circle) Bounds() Bounds { return BoundsAround(ci.Center, ci.Radius) }

// Area returns the spherical cap area.
func (ci Circle) Area() Area {
	R := float64(EarthRadius)
	return Area(2 * math.Pi * R * R * (1 - math.Cos(float64(ci.Radius/EarthRadius))))
}

// Circumference returns the length of the circle's edge on the sphere.
func (ci Circle) Circumference() Distance {
	return Distance(2*math.Pi*math.Sin(float64(ci.Radius/EarthRadius))) * EarthRadius
}
