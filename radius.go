package coordinatex

import "sort"

// WithinRadius reports whether p is within r of center (inclusive).
func WithinRadius(center, p Coordinate, r Distance) bool {
	return DistanceBetween(center, p) <= r
}

// WithinRadiusFilter returns the points within r of center, preserving order.
func WithinRadiusFilter(center Coordinate, points []Coordinate, r Distance) []Coordinate {
	var out []Coordinate
	for _, p := range points {
		if WithinRadius(center, p, r) {
			out = append(out, p)
		}
	}
	return out
}

// Nearest returns the closest point to origin, its index, and false if points is empty.
// It is a linear scan; use a spatial index for large datasets.
func Nearest(origin Coordinate, points []Coordinate) (Coordinate, int, bool) {
	best, bestD := -1, Distance(0)
	for i, p := range points {
		if d := DistanceBetween(origin, p); best < 0 || d < bestD {
			best, bestD = i, d
		}
	}
	if best < 0 {
		return Coordinate{}, -1, false
	}
	return points[best], best, true
}

// NearestN returns up to n points closest to origin, nearest first. Input is not modified.
func NearestN(origin Coordinate, points []Coordinate, n int) []Coordinate {
	if n <= 0 || len(points) == 0 {
		return nil
	}
	cp := append([]Coordinate(nil), points...)
	SortByDistance(origin, cp)
	if n > len(cp) {
		n = len(cp)
	}
	return cp[:n]
}

// SortByDistance sorts points in place by distance from origin (stable).
func SortByDistance(origin Coordinate, points []Coordinate) {
	type item struct {
		c Coordinate
		d Distance
	}
	items := make([]item, len(points))
	for i, p := range points {
		items[i] = item{p, DistanceBetween(origin, p)}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].d < items[j].d })
	for i := range items {
		points[i] = items[i].c
	}
}

// DistanceMatrix returns m[i][j] = DistanceBetween(from[i], to[j]).
func DistanceMatrix(from, to []Coordinate) [][]Distance {
	m := make([][]Distance, len(from))
	for i, f := range from {
		m[i] = make([]Distance, len(to))
		for j, t := range to {
			m[i][j] = DistanceBetween(f, t)
		}
	}
	return m
}
