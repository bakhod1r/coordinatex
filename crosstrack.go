package coordinatex

import "math"

// CrossTrackDistance returns the signed distance from p to the great circle
// through pathStart and pathEnd. Negative means left of the path, positive right.
func CrossTrackDistance(p, pathStart, pathEnd Coordinate) Distance {
	δ13 := float64(DistanceBetween(pathStart, p) / EarthRadius)
	θ13 := Bearing(pathStart, p).Radians()
	θ12 := Bearing(pathStart, pathEnd).Radians()
	return Distance(math.Asin(math.Sin(δ13)*math.Sin(θ13-θ12))) * EarthRadius
}

// AlongTrackDistance returns the distance from pathStart to the point on the
// path closest to p. Negative when p lies behind pathStart.
func AlongTrackDistance(p, pathStart, pathEnd Coordinate) Distance {
	δ13 := float64(DistanceBetween(pathStart, p) / EarthRadius)
	θ13 := Bearing(pathStart, p).Radians()
	θ12 := Bearing(pathStart, pathEnd).Radians()
	δxt := math.Asin(math.Sin(δ13) * math.Sin(θ13-θ12))
	c := math.Max(-1, math.Min(1, math.Cos(δ13)/math.Cos(δxt)))
	return Distance(math.Copysign(math.Acos(c), math.Cos(θ12-θ13))) * EarthRadius
}

type vec3 [3]float64

func toVec(c Coordinate) vec3 {
	φ, λ := toRad(c.Lat), toRad(c.Lng)
	return vec3{math.Cos(φ) * math.Cos(λ), math.Cos(φ) * math.Sin(λ), math.Sin(φ)}
}

func (a vec3) cross(b vec3) vec3 {
	return vec3{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

func (a vec3) norm() float64 { return math.Sqrt(a[0]*a[0] + a[1]*a[1] + a[2]*a[2]) }

func (a vec3) angleTo(b vec3) float64 {
	return math.Atan2(a.cross(b).norm(), a[0]*b[0]+a[1]*b[1]+a[2]*b[2])
}

func (a vec3) toCoordinate() Coordinate {
	return Coordinate{Lat: toDeg(math.Atan2(a[2], math.Hypot(a[0], a[1]))), Lng: wrapLng(toDeg(math.Atan2(a[1], a[0])))}
}

// SegmentIntersection returns the point where the great-circle segments a1-a2
// and b1-b2 cross. ok is false when they do not cross or lie on the same great circle.
func SegmentIntersection(a1, a2, b1, b2 Coordinate) (point Coordinate, ok bool) {
	va1, va2, vb1, vb2 := toVec(a1), toVec(a2), toVec(b1), toVec(b2)
	i := va1.cross(va2).cross(vb1.cross(vb2))
	n := i.norm()
	if n < 1e-12 {
		return Coordinate{}, false
	}
	onArc := func(p, s, e vec3) bool { return math.Abs(s.angleTo(p)+p.angleTo(e)-s.angleTo(e)) < 1e-9 }
	for _, sign := range []float64{1, -1} {
		p := vec3{sign * i[0] / n, sign * i[1] / n, sign * i[2] / n}
		if onArc(p, va1, va2) && onArc(p, vb1, vb2) {
			return p.toCoordinate(), true
		}
	}
	return Coordinate{}, false
}
