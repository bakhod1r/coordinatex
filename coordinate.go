package coordinatex

import (
	"fmt"
	"math"
)

// Coordinate is a point on Earth in decimal degrees.
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// New returns a validated Coordinate.
func New(lat, lng float64) (Coordinate, error) {
	c := Coordinate{Lat: lat, Lng: lng}
	if err := c.Validate(); err != nil {
		return Coordinate{}, err
	}
	return c, nil
}

// MustNew is like New but panics on invalid input. Intended for constants and tests.
func MustNew(lat, lng float64) Coordinate {
	c, err := New(lat, lng)
	if err != nil {
		panic(err)
	}
	return c
}

// Validate reports why c is not a valid coordinate, or nil.
func (c Coordinate) Validate() error {
	if !isFinite(c.Lat) || !isFinite(c.Lng) {
		return fmt.Errorf("%w: (%v, %v)", ErrNaN, c.Lat, c.Lng)
	}
	if c.Lat < -90 || c.Lat > 90 {
		return fmt.Errorf("%w: %v", ErrInvalidLatitude, c.Lat)
	}
	if c.Lng < -180 || c.Lng > 180 {
		return fmt.Errorf("%w: %v", ErrInvalidLongitude, c.Lng)
	}
	return nil
}

// IsValid reports whether c passes Validate.
func (c Coordinate) IsValid() bool { return c.Validate() == nil }

// IsZero reports whether c is the zero value (0, 0), often a sign of missing data.
func (c Coordinate) IsZero() bool { return c.Lat == 0 && c.Lng == 0 }

// Equal reports whether both components differ by at most tolerance degrees.
func (c Coordinate) Equal(o Coordinate, tolerance float64) bool {
	return math.Abs(c.Lat-o.Lat) <= tolerance && math.Abs(c.Lng-o.Lng) <= tolerance
}

// Normalize clamps latitude to [-90, 90] and wraps longitude to [-180, 180).
func (c Coordinate) Normalize() Coordinate {
	return Coordinate{Lat: math.Max(-90, math.Min(90, c.Lat)), Lng: wrapLng(c.Lng)}
}

// Antipode returns the point on the opposite side of the Earth.
func (c Coordinate) Antipode() Coordinate {
	return Coordinate{Lat: -c.Lat, Lng: wrapLng(c.Lng + 180)}
}

// String returns "lat,lng" with 6 decimal places (~0.1 m).
func (c Coordinate) String() string { return fmt.Sprintf("%.6f,%.6f", c.Lat, c.Lng) }

func isFinite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }

// wrapLng wraps a longitude to [-180, 180).
func wrapLng(lng float64) float64 {
	l := math.Mod(lng+180, 360)
	if l < 0 {
		l += 360
	}
	return l - 180
}

func toRad(deg float64) float64 { return deg * math.Pi / 180 }
func toDeg(rad float64) float64 { return rad * 180 / math.Pi }

// Round rounds both components to prec decimal places.
func (c Coordinate) Round(prec int) Coordinate {
	f := math.Pow10(prec)
	return Coordinate{Lat: math.Round(c.Lat*f) / f, Lng: math.Round(c.Lng*f) / f}
}

// MarshalText encodes as "lat,lng" (shortest exact decimal form).
func (c Coordinate) MarshalText() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return []byte(c.FormatDecimal(-1)), nil
}

// UnmarshalText accepts any format understood by Parse.
func (c *Coordinate) UnmarshalText(b []byte) error {
	p, err := Parse(string(b))
	if err != nil {
		return err
	}
	*c = p
	return nil
}
