package coordinatex

import "math"

// Destination returns the point reached by travelling distance d from start
// along the initial bearing, following a great circle.
func Destination(start Coordinate, d Distance, bearing Angle) Coordinate {
	δ := float64(d / EarthRadius)
	θ := bearing.Radians()
	φ1, λ1 := toRad(start.Lat), toRad(start.Lng)
	sinφ2 := math.Sin(φ1)*math.Cos(δ) + math.Cos(φ1)*math.Sin(δ)*math.Cos(θ)
	φ2 := math.Asin(sinφ2)
	λ2 := λ1 + math.Atan2(math.Sin(θ)*math.Sin(δ)*math.Cos(φ1), math.Cos(δ)-math.Sin(φ1)*sinφ2)
	return Coordinate{Lat: toDeg(φ2), Lng: wrapLng(toDeg(λ2))}
}

// Midpoint returns the great-circle midpoint between a and b.
func Midpoint(a, b Coordinate) Coordinate {
	φ1, λ1 := toRad(a.Lat), toRad(a.Lng)
	φ2 := toRad(b.Lat)
	Δλ := toRad(b.Lng - a.Lng)
	bx := math.Cos(φ2) * math.Cos(Δλ)
	by := math.Cos(φ2) * math.Sin(Δλ)
	φ3 := math.Atan2(math.Sin(φ1)+math.Sin(φ2), math.Hypot(math.Cos(φ1)+bx, by))
	λ3 := λ1 + math.Atan2(by, math.Cos(φ1)+bx)
	return Coordinate{Lat: toDeg(φ3), Lng: wrapLng(toDeg(λ3))}
}

// Interpolate returns the point at fraction (0 = a, 1 = b) along the great circle a -> b.
func Interpolate(a, b Coordinate, fraction float64) Coordinate {
	δ := float64(DistanceBetween(a, b) / EarthRadius)
	if δ == 0 {
		return a
	}
	φ1, λ1 := toRad(a.Lat), toRad(a.Lng)
	φ2, λ2 := toRad(b.Lat), toRad(b.Lng)
	A := math.Sin((1-fraction)*δ) / math.Sin(δ)
	B := math.Sin(fraction*δ) / math.Sin(δ)
	x := A*math.Cos(φ1)*math.Cos(λ1) + B*math.Cos(φ2)*math.Cos(λ2)
	y := A*math.Cos(φ1)*math.Sin(λ1) + B*math.Cos(φ2)*math.Sin(λ2)
	z := A*math.Sin(φ1) + B*math.Sin(φ2)
	return Coordinate{Lat: toDeg(math.Atan2(z, math.Hypot(x, y))), Lng: wrapLng(toDeg(math.Atan2(y, x)))}
}

// IntermediatePoints returns a, n evenly spaced great-circle points, and b (n+2 points).
func IntermediatePoints(a, b Coordinate, n int) []Coordinate {
	if n < 0 {
		n = 0
	}
	out := make([]Coordinate, n+2)
	for i := range out {
		out[i] = Interpolate(a, b, float64(i)/float64(n+1))
	}
	out[0], out[n+1] = a, b
	return out
}

// PointsEvery returns points every step along the great circle from a, ending with b.
// A non-positive step returns just a and b.
func PointsEvery(a, b Coordinate, step Distance) []Coordinate {
	total := DistanceBetween(a, b)
	if step <= 0 || total == 0 {
		return []Coordinate{a, b}
	}
	out := []Coordinate{a}
	for d := step; d < total; d += step {
		out = append(out, Interpolate(a, b, float64(d/total)))
	}
	return append(out, b)
}
