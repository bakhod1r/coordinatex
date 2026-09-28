package coordinatex

import (
	"fmt"
	"math"
)

// Bounds is a latitude/longitude rectangle. When MinLng > MaxLng the box
// crosses the antimeridian (180°).
type Bounds struct {
	MinLat float64 `json:"minLat"`
	MinLng float64 `json:"minLng"`
	MaxLat float64 `json:"maxLat"`
	MaxLng float64 `json:"maxLng"`
}

// NewBounds builds Bounds from south-west and north-east corners.
// sw.Lng > ne.Lng is allowed and means the box crosses the antimeridian.
func NewBounds(sw, ne Coordinate) (Bounds, error) {
	if err := sw.Validate(); err != nil {
		return Bounds{}, err
	}
	if err := ne.Validate(); err != nil {
		return Bounds{}, err
	}
	if sw.Lat > ne.Lat {
		return Bounds{}, fmt.Errorf("%w: south %v > north %v", ErrInvalidBounds, sw.Lat, ne.Lat)
	}
	return Bounds{MinLat: sw.Lat, MinLng: sw.Lng, MaxLat: ne.Lat, MaxLng: ne.Lng}, nil
}

// BoundsOf returns the smallest non-antimeridian-crossing box containing all points.
func BoundsOf(points ...Coordinate) (Bounds, error) {
	if len(points) == 0 {
		return Bounds{}, ErrEmpty
	}
	b := Bounds{MinLat: 90, MinLng: 180, MaxLat: -90, MaxLng: -180}
	for _, p := range points {
		if err := p.Validate(); err != nil {
			return Bounds{}, err
		}
		b.MinLat, b.MaxLat = math.Min(b.MinLat, p.Lat), math.Max(b.MaxLat, p.Lat)
		b.MinLng, b.MaxLng = math.Min(b.MinLng, p.Lng), math.Max(b.MaxLng, p.Lng)
	}
	return b, nil
}

// BoundsAround returns a box that encloses the circle of the given radius.
// Useful for a cheap index pre-filter (WHERE lat BETWEEN ... AND lng BETWEEN ...)
// followed by an exact WithinRadius check.
func BoundsAround(center Coordinate, radius Distance) Bounds {
	return Bounds{MinLat: center.Lat, MinLng: center.Lng, MaxLat: center.Lat, MaxLng: center.Lng}.Expand(radius)
}

// Expand grows the box by d in every direction.
func (b Bounds) Expand(d Distance) Bounds {
	δ := toDeg(float64(d / EarthRadius))
	out := Bounds{MinLat: b.MinLat - δ, MaxLat: b.MaxLat + δ, MinLng: b.MinLng, MaxLng: b.MaxLng}
	if out.MinLat <= -90 || out.MaxLat >= 90 {
		out.MinLat, out.MaxLat = math.Max(out.MinLat, -90), math.Min(out.MaxLat, 90)
		out.MinLng, out.MaxLng = -180, 180
		return out
	}
	// Widest longitude offset occurs at the most poleward latitude of the box.
	φ := toRad(math.Max(math.Abs(b.MinLat), math.Abs(b.MaxLat)))
	Δλ := toDeg(math.Asin(math.Min(1, math.Sin(toRad(δ))/math.Cos(φ))))
	if b.lngSpan()+2*Δλ >= 360 {
		out.MinLng, out.MaxLng = -180, 180
		return out
	}
	out.MinLng, out.MaxLng = wrapLng(b.MinLng-Δλ), b.MaxLng+Δλ
	if out.MaxLng > 180 {
		out.MaxLng -= 360
	}
	return out
}

// Contains reports whether c lies inside or on the edge of b.
func (b Bounds) Contains(c Coordinate) bool {
	if c.Lat < b.MinLat || c.Lat > b.MaxLat {
		return false
	}
	if b.CrossesAntimeridian() {
		return c.Lng >= b.MinLng || c.Lng <= b.MaxLng
	}
	return c.Lng >= b.MinLng && c.Lng <= b.MaxLng
}

// CrossesAntimeridian reports whether the box wraps across 180°.
func (b Bounds) CrossesAntimeridian() bool { return b.MinLng > b.MaxLng }

// Intersects reports whether the two boxes overlap.
func (b Bounds) Intersects(o Bounds) bool {
	if b.MaxLat < o.MinLat || o.MaxLat < b.MinLat {
		return false
	}
	for _, x := range b.lngIntervals() {
		for _, y := range o.lngIntervals() {
			if x[0] <= y[1] && y[0] <= x[1] {
				return true
			}
		}
	}
	return false
}

// Union returns the smallest box containing both. Antimeridian-crossing inputs
// are not merged optimally.
func (b Bounds) Union(o Bounds) Bounds {
	return Bounds{
		MinLat: math.Min(b.MinLat, o.MinLat), MinLng: math.Min(b.MinLng, o.MinLng),
		MaxLat: math.Max(b.MaxLat, o.MaxLat), MaxLng: math.Max(b.MaxLng, o.MaxLng),
	}
}

// Center returns the midpoint of the box in lat/lng space.
func (b Bounds) Center() Coordinate {
	return Coordinate{Lat: (b.MinLat + b.MaxLat) / 2, Lng: wrapLng(b.MinLng + b.lngSpan()/2)}
}

// Width is the east-west extent measured along the center latitude.
func (b Bounds) Width() Distance {
	return Distance(toRad(b.lngSpan())*math.Cos(toRad((b.MinLat+b.MaxLat)/2))) * EarthRadius
}

// Height is the north-south extent.
func (b Bounds) Height() Distance { return Distance(toRad(b.MaxLat-b.MinLat)) * EarthRadius }

// SouthWest returns the (MinLat, MinLng) corner.
func (b Bounds) SouthWest() Coordinate { return Coordinate{Lat: b.MinLat, Lng: b.MinLng} }

// NorthEast returns the (MaxLat, MaxLng) corner.
func (b Bounds) NorthEast() Coordinate { return Coordinate{Lat: b.MaxLat, Lng: b.MaxLng} }

func (b Bounds) lngSpan() float64 {
	if b.CrossesAntimeridian() {
		return b.MaxLng + 360 - b.MinLng
	}
	return b.MaxLng - b.MinLng
}

func (b Bounds) lngIntervals() [][2]float64 {
	if b.CrossesAntimeridian() {
		return [][2]float64{{b.MinLng, 180}, {-180, b.MaxLng}}
	}
	return [][2]float64{{b.MinLng, b.MaxLng}}
}

// Intersection returns the overlapping box and whether one exists.
// For antimeridian-crossing inputs the first overlapping longitude interval is returned.
func (b Bounds) Intersection(o Bounds) (Bounds, bool) {
	minLat, maxLat := math.Max(b.MinLat, o.MinLat), math.Min(b.MaxLat, o.MaxLat)
	if minLat > maxLat {
		return Bounds{}, false
	}
	for _, x := range b.lngIntervals() {
		for _, y := range o.lngIntervals() {
			if lo, hi := math.Max(x[0], y[0]), math.Min(x[1], y[1]); lo <= hi {
				return Bounds{MinLat: minLat, MinLng: lo, MaxLat: maxLat, MaxLng: hi}, true
			}
		}
	}
	return Bounds{}, false
}

// Area returns the spherical surface area of the box.
func (b Bounds) Area() Area {
	R := float64(EarthRadius)
	return Area(R * R * toRad(b.lngSpan()) * (math.Sin(toRad(b.MaxLat)) - math.Sin(toRad(b.MinLat))))
}

// Perimeter returns the length of the two meridian sides plus the two parallels.
func (b Bounds) Perimeter() Distance {
	λ := toRad(b.lngSpan())
	par := Distance(λ*(math.Cos(toRad(b.MinLat))+math.Cos(toRad(b.MaxLat)))) * EarthRadius
	return 2*b.Height() + par
}

// Split divides the box into rows x cols cells, row-major from the south-west.
// Useful for grid partitioning and parallel spatial work.
func (b Bounds) Split(rows, cols int) []Bounds {
	if rows < 1 || cols < 1 {
		return nil
	}
	dLat, dLng := (b.MaxLat-b.MinLat)/float64(rows), b.lngSpan()/float64(cols)
	out := make([]Bounds, 0, rows*cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			lo, hi := b.MinLng+float64(c)*dLng, b.MinLng+float64(c+1)*dLng
			if lo >= 180 {
				lo, hi = lo-360, hi-360
			}
			if c == cols-1 {
				hi = b.MaxLng
			}
			cell := Bounds{MinLat: b.MinLat + float64(r)*dLat, MinLng: lo, MaxLat: b.MinLat + float64(r+1)*dLat, MaxLng: hi}
			if r == rows-1 {
				cell.MaxLat = b.MaxLat
			}
			out = append(out, cell)
		}
	}
	return out
}
