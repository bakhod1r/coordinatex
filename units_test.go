package coordinatex

import (
	"math"
	"testing"
)

func TestDistanceUnits(t *testing.T) {
	d := 1609.344 * Meter
	near(t, "Miles", d.Miles(), 1, 1e-12)
	near(t, "Meters", d.Meters(), 1609.344, 1e-12)
	near(t, "Kilometers", (2 * Kilometer).Kilometers(), 2, 1e-12)
	near(t, "Feet", Foot.Feet(), 1, 1e-12)
	near(t, "NauticalMiles", (3 * NauticalMile).NauticalMiles(), 3, 1e-12)
	if s := (1500 * Meter).String(); s != "1.50 km" {
		t.Fatal(s)
	}
	if s := (12.346 * Meter).String(); s != "12.35 m" {
		t.Fatal(s)
	}
}

func TestAngle(t *testing.T) {
	near(t, "Radians", Angle(180).Radians(), math.Pi, 1e-12)
	near(t, "Degrees", Angle(45).Degrees(), 45, 0)
	near(t, "Normalize neg", float64(Angle(-90).Normalize()), 270, 1e-12)
	near(t, "Normalize big", float64(Angle(725).Normalize()), 5, 1e-12)
	near(t, "Normalize 360", float64(Angle(360).Normalize()), 0, 1e-12)
}
