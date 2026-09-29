package coordinatex

import (
	"math"
	"sort"
	"time"
)

// ContainsSpherical reports whether c is inside the polygon using great-circle
// edges (winding of bearings). Unlike Contains it handles polygons that cross
// the antimeridian or surround a pole. Points on a vertex count as inside.
// Polygons must be smaller than a hemisphere.
func (p Polygon) ContainsSpherical(c Coordinate) bool {
	r := p.ring()
	if len(r) < 3 {
		return false
	}
	// Winding cannot tell c from its antipode; require c on the polygon's hemisphere.
	var centroid vec3
	for _, v := range r {
		w := toVec(v)
		centroid = vec3{centroid[0] + w[0], centroid[1] + w[1], centroid[2] + w[2]}
	}
	if cv := toVec(c); cv[0]*centroid[0]+cv[1]*centroid[1]+cv[2]*centroid[2] <= 0 {
		return false
	}
	var sum Angle
	for i := range r {
		if DistanceBetween(c, r[i]) < 1e-6 {
			return true
		}
		sum += AngleDifference(Bearing(c, r[i]), Bearing(c, r[(i+1)%len(r)]))
	}
	return math.Abs(float64(sum)) > 180
}

// DistanceTo returns 0 when c is inside (spherical test), otherwise the
// distance to the nearest edge.
func (p Polygon) DistanceTo(c Coordinate) Distance {
	r := p.ring()
	if len(r) == 0 {
		return Distance(math.Inf(1))
	}
	if p.ContainsSpherical(c) {
		return 0
	}
	best := Distance(math.Inf(1))
	for i := range r {
		if d := DistanceBetween(c, closestOnSegment(c, r[i], r[(i+1)%len(r)])); d < best {
			best = d
		}
	}
	return best
}

// DistanceTo returns 0 inside the fence, otherwise distance to its edge.
func (f PolygonFence) DistanceTo(c Coordinate) Distance { return f.Polygon.DistanceTo(c) }

// DistanceTo returns 0 inside the fence, otherwise distance to its edge.
func (f CircleFence) DistanceTo(c Coordinate) Distance {
	return max(0, Circle(f).DistanceTo(c))
}

// DistanceTo returns 0 inside the box, otherwise the distance to its edge
// (edges approximated by great circles between corners).
func (f RectangleFence) DistanceTo(c Coordinate) Distance {
	if f.Bounds.Contains(c) {
		return 0
	}
	b := f.Bounds
	return Polygon{{b.MinLat, b.MinLng}, {b.MinLat, b.MaxLng}, {b.MaxLat, b.MaxLng}, {b.MaxLat, b.MinLng}}.DistanceTo(c)
}

// SphericalPolygonFence uses ContainsSpherical; safe across the antimeridian and poles.
type SphericalPolygonFence struct{ Polygon Polygon }

func (f SphericalPolygonFence) Contains(c Coordinate) bool { return f.Polygon.ContainsSpherical(c) }

// CorridorFence contains points within Width of Route (a buffer around a line).
type CorridorFence struct {
	Route Polyline
	Width Distance
}

func (f CorridorFence) Contains(c Coordinate) bool { return !f.Route.Deviates(c, f.Width) }

// Polygon approximates the circle with n vertices (minimum 3), clockwise from north.
func (ci Circle) Polygon(n int) Polygon {
	if n < 3 {
		n = 3
	}
	out := make(Polygon, n)
	for i := range out {
		out[i] = Destination(ci.Center, ci.Radius, Angle(360*float64(i)/float64(n)))
	}
	return out
}

// ConvexHull returns the counter-clockwise convex hull of the points (planar in
// lat/lng space). Fewer than 3 non-collinear points return the extreme points.
func ConvexHull(points []Coordinate) Polygon {
	pts := append([]Coordinate(nil), points...)
	sort.Slice(pts, func(i, j int) bool {
		if pts[i].Lng != pts[j].Lng {
			return pts[i].Lng < pts[j].Lng
		}
		return pts[i].Lat < pts[j].Lat
	})
	uniq := pts[:0]
	for i, p := range pts {
		if i == 0 || p != pts[i-1] {
			uniq = append(uniq, p)
		}
	}
	if len(uniq) < 3 {
		return Polygon(uniq)
	}
	hull := make([]Coordinate, 0, 2*len(uniq))
	for pass := 0; pass < 2; pass++ {
		start := len(hull)
		for _, p := range uniq {
			for len(hull) >= start+2 && orient(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
				hull = hull[:len(hull)-1]
			}
			hull = append(hull, p)
		}
		hull = hull[:len(hull)-1]
		for i, j := 0, len(uniq)-1; i < j; i, j = i+1, j-1 {
			uniq[i], uniq[j] = uniq[j], uniq[i]
		}
	}
	return Polygon(hull)
}

// Noise is the DBSCAN label for points that belong to no cluster.
const Noise = -1

// DBSCAN clusters points by density: a cluster is a set of points each with at
// least minPts neighbours within eps. Returns a label per point (Noise or
// 0..n-1) and the cluster count.
func DBSCAN(points []Coordinate, eps Distance, minPts int) (labels []int, clusters int) {
	if len(points) == 0 {
		return nil, 0
	}
	idx := NewGridIndex[int](eps)
	for i, p := range points {
		idx.Insert(i, p)
	}
	const unvisited = -2
	labels = make([]int, len(points))
	for i := range labels {
		labels[i] = unvisited
	}
	for i := range points {
		if labels[i] != unvisited {
			continue
		}
		nb := idx.WithinRadius(points[i], eps)
		if len(nb) < minPts {
			labels[i] = Noise
			continue
		}
		labels[i] = clusters
		queue := nb
		for len(queue) > 0 {
			j := queue[0]
			queue = queue[1:]
			if labels[j] == Noise {
				labels[j] = clusters
			}
			if labels[j] != unvisited {
				continue
			}
			labels[j] = clusters
			if nb2 := idx.WithinRadius(points[j], eps); len(nb2) >= minPts {
				queue = append(queue, nb2...)
			}
		}
		clusters++
	}
	return labels, clusters
}

// Stop is a period where the track stayed within a radius.
type Stop struct {
	Center     Coordinate
	Start, End time.Time
	Points     int
	first      int
}

// Duration is End - Start.
func (s Stop) Duration() time.Duration { return s.End.Sub(s.Start) }

// Stops finds periods where consecutive fixes stay within radius of the
// period's first fix for at least minDuration.
func (t Track) Stops(radius Distance, minDuration time.Duration) []Stop {
	var out []Stop
	for i := 0; i < len(t); {
		j := i + 1
		for j < len(t) && WithinRadius(t[i].Coordinate, t[j].Coordinate, radius) {
			j++
		}
		if t[j-1].Time.Sub(t[i].Time) >= minDuration {
			var c Coordinate
			for _, p := range t[i:j] {
				c.Lat += p.Lat
				c.Lng += p.Lng
			}
			n := float64(j - i)
			out = append(out, Stop{Center: Coordinate{Lat: c.Lat / n, Lng: c.Lng / n}, Start: t[i].Time, End: t[j-1].Time, Points: j - i, first: i})
			i = j
			continue
		}
		i++
	}
	return out
}

// Trips splits the track into the moving parts between Stops.
func (t Track) Trips(radius Distance, minDuration time.Duration) []Track {
	var out []Track
	from := 0
	for _, s := range t.Stops(radius, minDuration) {
		if s.first > from {
			out = append(out, t[from:s.first])
		}
		from = s.first + s.Points
	}
	if from < len(t) {
		out = append(out, t[from:])
	}
	return out
}
