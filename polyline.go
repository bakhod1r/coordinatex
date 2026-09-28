package coordinatex

import (
	"fmt"
	"math"
	"strings"
)

// Polyline is an ordered sequence of points, such as a GPS track or route.
type Polyline []Coordinate

// Length returns the sum of great-circle segment lengths.
func (l Polyline) Length() Distance {
	var d Distance
	for i := 1; i < len(l); i++ {
		d += DistanceBetween(l[i-1], l[i])
	}
	return d
}

// Bounds returns the bounding box of the points.
func (l Polyline) Bounds() (Bounds, error) { return BoundsOf(l...) }

// NearestPoint returns the closest point on the line to c and its distance.
// ok is false for an empty line.
func (l Polyline) NearestPoint(c Coordinate) (point Coordinate, d Distance, ok bool) {
	switch len(l) {
	case 0:
		return Coordinate{}, 0, false
	case 1:
		return l[0], DistanceBetween(c, l[0]), true
	}
	d = Distance(math.Inf(1))
	for i := 1; i < len(l); i++ {
		p := closestOnSegment(c, l[i-1], l[i])
		if pd := DistanceBetween(c, p); pd < d {
			point, d = p, pd
		}
	}
	return point, d, true
}

func closestOnSegment(c, a, b Coordinate) Coordinate {
	seg := DistanceBetween(a, b)
	if seg == 0 {
		return a
	}
	along := AlongTrackDistance(c, a, b)
	switch {
	case along <= 0:
		return a
	case along >= seg:
		return b
	}
	return Destination(a, along, Bearing(a, b))
}

// Simplify reduces points with the Ramer–Douglas–Peucker algorithm, keeping
// every point farther than tolerance from the simplified line.
func (l Polyline) Simplify(tolerance Distance) Polyline {
	if len(l) < 3 || tolerance <= 0 {
		return append(Polyline(nil), l...)
	}
	keep := make([]bool, len(l))
	keep[0], keep[len(l)-1] = true, true
	var rdp func(lo, hi int)
	rdp = func(lo, hi int) {
		maxD, idx := Distance(-1), -1
		for i := lo + 1; i < hi; i++ {
			if d := DistanceBetween(l[i], closestOnSegment(l[i], l[lo], l[hi])); d > maxD {
				maxD, idx = d, i
			}
		}
		if idx >= 0 && maxD > tolerance {
			keep[idx] = true
			rdp(lo, idx)
			rdp(idx, hi)
		}
	}
	rdp(0, len(l)-1)
	out := make(Polyline, 0, len(l))
	for i, k := range keep {
		if k {
			out = append(out, l[i])
		}
	}
	return out
}

// EncodePolyline encodes points in Google's encoded polyline format.
// precision is 5 for Google Maps, 6 for OSRM/Valhalla.
func EncodePolyline(l Polyline, precision int) string {
	f := math.Pow10(precision)
	var sb strings.Builder
	var pLat, pLng int64
	for _, c := range l {
		lat, lng := int64(math.Round(c.Lat*f)), int64(math.Round(c.Lng*f))
		encodeSigned(&sb, lat-pLat)
		encodeSigned(&sb, lng-pLng)
		pLat, pLng = lat, lng
	}
	return sb.String()
}

func encodeSigned(sb *strings.Builder, v int64) {
	u := uint64(v) << 1
	if v < 0 {
		u = ^u
	}
	for u >= 0x20 {
		sb.WriteByte(byte(0x20|u&0x1f) + 63)
		u >>= 5
	}
	sb.WriteByte(byte(u) + 63)
}

// DecodePolyline decodes a Google encoded polyline.
func DecodePolyline(s string, precision int) (Polyline, error) {
	f := math.Pow10(precision)
	var out Polyline
	var lat, lng int64
	for i := 0; i < len(s); {
		dLat, n, err := decodeSigned(s, i)
		if err != nil {
			return nil, err
		}
		i = n
		if i >= len(s) {
			return nil, fmt.Errorf("%w: missing longitude", ErrInvalidPolyline)
		}
		dLng, n, err := decodeSigned(s, i)
		if err != nil {
			return nil, err
		}
		i = n
		lat, lng = lat+dLat, lng+dLng
		c := Coordinate{Lat: float64(lat) / f, Lng: float64(lng) / f}
		if err := c.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPolyline, err)
		}
		out = append(out, c)
	}
	return out, nil
}

func decodeSigned(s string, i int) (int64, int, error) {
	var u uint64
	for shift := uint(0); ; shift += 5 {
		if i >= len(s) || shift > 60 {
			return 0, i, fmt.Errorf("%w: truncated", ErrInvalidPolyline)
		}
		b := int(s[i]) - 63
		i++
		if b < 0 || b > 63 {
			return 0, i, fmt.Errorf("%w: bad byte", ErrInvalidPolyline)
		}
		u |= uint64(b&0x1f) << shift
		if b < 0x20 {
			break
		}
	}
	v := int64(u >> 1)
	if u&1 != 0 {
		v = ^v
	}
	return v, i, nil
}

// Reverse returns a reversed copy.
func (l Polyline) Reverse() Polyline {
	out := make(Polyline, len(l))
	for i, c := range l {
		out[len(l)-1-i] = c
	}
	return out
}

// SegmentLengths returns the length of each segment (len(l)-1 values).
func (l Polyline) SegmentLengths() []Distance {
	if len(l) < 2 {
		return nil
	}
	out := make([]Distance, len(l)-1)
	for i := range out {
		out[i] = DistanceBetween(l[i], l[i+1])
	}
	return out
}

// Bearings returns the initial bearing of each segment.
func (l Polyline) Bearings() []Angle {
	if len(l) < 2 {
		return nil
	}
	out := make([]Angle, len(l)-1)
	for i := range out {
		out[i] = Bearing(l[i], l[i+1])
	}
	return out
}

// PointAt returns the point at distance d along the line, clamped to its ends.
func (l Polyline) PointAt(d Distance) (Coordinate, bool) {
	if len(l) == 0 {
		return Coordinate{}, false
	}
	if d <= 0 {
		return l[0], true
	}
	for i := 1; i < len(l); i++ {
		seg := DistanceBetween(l[i-1], l[i])
		if d <= seg && seg > 0 {
			return Interpolate(l[i-1], l[i], float64(d/seg)), true
		}
		d -= seg
	}
	return l[len(l)-1], true
}

// Deviates reports whether c is farther than tolerance from the route.
// An empty route always deviates.
func (l Polyline) Deviates(c Coordinate, tolerance Distance) bool {
	_, d, ok := l.NearestPoint(c)
	return !ok || d > tolerance
}
