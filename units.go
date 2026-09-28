package coordinatex

import (
	"fmt"
	"math"
)

// Distance is a length in meters.
type Distance float64

// Common distance units.
const (
	Meter        Distance = 1
	Kilometer    Distance = 1000
	Mile         Distance = 1609.344
	Foot         Distance = 0.3048
	NauticalMile Distance = 1852
	Centimeter   Distance = 0.01
	Yard         Distance = 0.9144
)

// EarthRadius is the IUGG mean Earth radius used by spherical calculations.
const EarthRadius Distance = 6371008.8

func (d Distance) Meters() float64        { return float64(d) }
func (d Distance) Kilometers() float64    { return float64(d / Kilometer) }
func (d Distance) Miles() float64         { return float64(d / Mile) }
func (d Distance) Feet() float64          { return float64(d / Foot) }
func (d Distance) NauticalMiles() float64 { return float64(d / NauticalMile) }
func (d Distance) Centimeters() float64   { return float64(d / Centimeter) }
func (d Distance) Yards() float64         { return float64(d / Yard) }

// String formats as meters below 1 km, kilometers otherwise.
func (d Distance) String() string {
	if math.Abs(float64(d)) >= float64(Kilometer) {
		return fmt.Sprintf("%.2f km", d.Kilometers())
	}
	return fmt.Sprintf("%.2f m", d.Meters())
}

// Angle is an angle in degrees.
type Angle float64

func (a Angle) Degrees() float64 { return float64(a) }
func (a Angle) Radians() float64 { return toRad(float64(a)) }

// Normalize wraps the angle to [0, 360).
func (a Angle) Normalize() Angle {
	n := math.Mod(float64(a), 360)
	if n < 0 {
		n += 360
	}
	return Angle(n)
}

// Reverse returns the opposite direction (back bearing).
func (a Angle) Reverse() Angle { return (a + 180).Normalize() }

// AngleDifference returns the smallest signed rotation from a to b in (-180, 180].
// Positive is clockwise.
func AngleDifference(a, b Angle) Angle {
	d := math.Mod(float64(b-a)+180, 360)
	if d < 0 {
		d += 360
	}
	d -= 180
	if d == -180 {
		d = 180
	}
	return Angle(d)
}

// Area is a surface area in square meters.
type Area float64

// Common area units.
const (
	SquareMeter     Area = 1
	Hectare         Area = 1e4
	SquareKilometer Area = 1e6
	Acre            Area = 4046.8564224
)

func (a Area) SquareMeters() float64     { return float64(a) }
func (a Area) SquareKilometers() float64 { return float64(a / SquareKilometer) }
func (a Area) Hectares() float64         { return float64(a / Hectare) }
func (a Area) Acres() float64            { return float64(a / Acre) }

// String formats as m² below 1 km², km² otherwise.
func (a Area) String() string {
	if math.Abs(float64(a)) >= float64(SquareKilometer) {
		return fmt.Sprintf("%.2f km²", a.SquareKilometers())
	}
	return fmt.Sprintf("%.2f m²", a.SquareMeters())
}

// Speed is a velocity in meters per second.
type Speed float64

func (s Speed) MetersPerSecond() float64   { return float64(s) }
func (s Speed) KilometersPerHour() float64 { return float64(s) * 3.6 }
func (s Speed) MilesPerHour() float64      { return float64(s) * 3600 / float64(Mile) }
func (s Speed) Knots() float64             { return float64(s) * 3600 / float64(NauticalMile) }
