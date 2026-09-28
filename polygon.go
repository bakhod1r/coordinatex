package coordinatex

import (
	"fmt"
	"math"
)

// Polygon is a single ring of vertices. Closing the ring (repeating the first
// vertex at the end) is optional.
type Polygon []Coordinate

// ring returns vertices without the closing duplicate.
func (p Polygon) ring() []Coordinate {
	if n := len(p); n > 1 && p[0] == p[n-1] {
		return p[:n-1]
	}
	return p
}

// Validate checks vertex count and every vertex.
func (p Polygon) Validate() error {
	r := p.ring()
	if len(r) < 3 {
		return fmt.Errorf("%w: got %d", ErrInvalidPolygon, len(r))
	}
	for _, c := range r {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Contains reports whether c is inside the polygon (ray casting in lat/lng
// space; suitable for fences that do not span the antimeridian or poles).
func (p Polygon) Contains(c Coordinate) bool {
	r := p.ring()
	if len(r) < 3 {
		return false
	}
	inside := false
	for i, j := 0, len(r)-1; i < len(r); j, i = i, i+1 {
		a, b := r[i], r[j]
		if (a.Lat > c.Lat) != (b.Lat > c.Lat) &&
			c.Lng < (b.Lng-a.Lng)*(c.Lat-a.Lat)/(b.Lat-a.Lat)+a.Lng {
			inside = !inside
		}
	}
	return inside
}

// Area returns the spherical area.
func (p Polygon) Area() Area {
	r := p.ring()
	if len(r) < 3 {
		return 0
	}
	var sum float64
	for i := range r {
		a, b := r[i], r[(i+1)%len(r)]
		sum += toRad(b.Lng-a.Lng) * (2 + math.Sin(toRad(a.Lat)) + math.Sin(toRad(b.Lat)))
	}
	R := float64(EarthRadius)
	return Area(math.Abs(sum) * R * R / 2)
}

// Perimeter returns the length of the closed ring.
func (p Polygon) Perimeter() Distance {
	r := p.ring()
	if len(r) < 2 {
		return 0
	}
	return Polyline(append(append([]Coordinate{}, r...), r[0])).Length()
}

// signedArea is the planar shoelace sum in (lng, lat) space.
func (p Polygon) signedArea() float64 {
	r := p.ring()
	var s float64
	for i := range r {
		a, b := r[i], r[(i+1)%len(r)]
		s += a.Lng*b.Lat - b.Lng*a.Lat
	}
	return s / 2
}

// IsClockwise reports whether vertices wind clockwise as seen on a north-up map.
func (p Polygon) IsClockwise() bool { return p.signedArea() < 0 }

// Centroid returns the planar area-weighted centroid in lat/lng space, or the
// vertex average for degenerate (zero-area) polygons.
func (p Polygon) Centroid() Coordinate {
	r := p.ring()
	if len(r) == 0 {
		return Coordinate{}
	}
	A := p.signedArea()
	if A == 0 {
		var c Coordinate
		for _, v := range r {
			c.Lat += v.Lat
			c.Lng += v.Lng
		}
		return Coordinate{Lat: c.Lat / float64(len(r)), Lng: c.Lng / float64(len(r))}
	}
	var cx, cy float64
	for i := range r {
		a, b := r[i], r[(i+1)%len(r)]
		f := a.Lng*b.Lat - b.Lng*a.Lat
		cx += (a.Lng + b.Lng) * f
		cy += (a.Lat + b.Lat) * f
	}
	return Coordinate{Lat: cy / (6 * A), Lng: cx / (6 * A)}
}

// Bounds returns the bounding box of the vertices.
func (p Polygon) Bounds() (Bounds, error) { return BoundsOf(p...) }

// IsSimple reports whether no two non-adjacent edges intersect (planar test in lat/lng space).
func (p Polygon) IsSimple() bool {
	r := p.ring()
	n := len(r)
	for i := 0; i < n; i++ {
		a1, a2 := r[i], r[(i+1)%n]
		for j := i + 1; j < n; j++ {
			if j == i+1 || (i == 0 && j == n-1) {
				continue // adjacent edges share a vertex
			}
			if segmentsIntersect(a1, a2, r[j], r[(j+1)%n]) {
				return false
			}
		}
	}
	return true
}

func orient(a, b, c Coordinate) float64 {
	return (b.Lng-a.Lng)*(c.Lat-a.Lat) - (b.Lat-a.Lat)*(c.Lng-a.Lng)
}

func segmentsIntersect(p1, p2, q1, q2 Coordinate) bool {
	d1, d2 := orient(q1, q2, p1), orient(q1, q2, p2)
	d3, d4 := orient(p1, p2, q1), orient(p1, p2, q2)
	if ((d1 > 0) != (d2 > 0)) && ((d3 > 0) != (d4 > 0)) && d1 != 0 && d2 != 0 && d3 != 0 && d4 != 0 {
		return true
	}
	on := func(a, b, c Coordinate) bool {
		return math.Min(a.Lng, b.Lng) <= c.Lng && c.Lng <= math.Max(a.Lng, b.Lng) &&
			math.Min(a.Lat, b.Lat) <= c.Lat && c.Lat <= math.Max(a.Lat, b.Lat)
	}
	return (d1 == 0 && on(q1, q2, p1)) || (d2 == 0 && on(q1, q2, p2)) ||
		(d3 == 0 && on(p1, p2, q1)) || (d4 == 0 && on(p1, p2, q2))
}

func (p Polygon) edgesCross(o Polygon) bool {
	r, s := p.ring(), o.ring()
	for i := range r {
		for j := range s {
			if segmentsIntersect(r[i], r[(i+1)%len(r)], s[j], s[(j+1)%len(s)]) {
				return true
			}
		}
	}
	return false
}

// Intersects reports whether the polygons share any area or boundary (planar).
func (p Polygon) Intersects(o Polygon) bool {
	if len(p.ring()) < 3 || len(o.ring()) < 3 {
		return false
	}
	return p.edgesCross(o) || p.Contains(o[0]) || o.Contains(p[0])
}

// ContainsPolygon reports whether o lies entirely inside p (planar).
func (p Polygon) ContainsPolygon(o Polygon) bool {
	if len(p.ring()) < 3 || len(o.ring()) < 3 || p.edgesCross(o) {
		return false
	}
	for _, c := range o.ring() {
		if !p.Contains(c) {
			return false
		}
	}
	return true
}

// Overlaps reports whether the polygons intersect but neither contains the other.
func (p Polygon) Overlaps(o Polygon) bool {
	return p.Intersects(o) && !p.ContainsPolygon(o) && !o.ContainsPolygon(p)
}

// PolygonWithHoles is an exterior ring with optional interior rings (holes).
type PolygonWithHoles struct {
	Outer Polygon
	Holes []Polygon
}

// Validate checks every ring.
func (p PolygonWithHoles) Validate() error {
	if err := p.Outer.Validate(); err != nil {
		return err
	}
	for _, h := range p.Holes {
		if err := h.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Contains reports whether c is inside the exterior and outside every hole.
func (p PolygonWithHoles) Contains(c Coordinate) bool {
	if !p.Outer.Contains(c) {
		return false
	}
	for _, h := range p.Holes {
		if h.Contains(c) {
			return false
		}
	}
	return true
}

// Area is the exterior area minus the holes.
func (p PolygonWithHoles) Area() Area {
	a := p.Outer.Area()
	for _, h := range p.Holes {
		a -= h.Area()
	}
	return a
}

// Bounds returns the bounding box of the exterior ring.
func (p PolygonWithHoles) Bounds() (Bounds, error) { return p.Outer.Bounds() }
