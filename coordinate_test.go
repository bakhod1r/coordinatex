package coordinatex

import (
	"errors"
	"math"
	"testing"
)

func TestNewValidates(t *testing.T) {
	tests := []struct {
		name     string
		lat, lng float64
		err      error
	}{
		{"valid", 41.3, 69.2, nil},
		{"lat edge", 90, 180, nil},
		{"lat edge neg", -90, -180, nil},
		{"lat high", 90.0001, 0, ErrInvalidLatitude},
		{"lat low", -91, 0, ErrInvalidLatitude},
		{"lng high", 0, 180.1, ErrInvalidLongitude},
		{"lng low", 0, -181, ErrInvalidLongitude},
		{"nan", math.NaN(), 0, ErrNaN},
		{"inf", 0, math.Inf(1), ErrNaN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.lat, tt.lng)
			if !errors.Is(err, tt.err) {
				t.Fatalf("err = %v, want %v", err, tt.err)
			}
			if tt.err == nil && (c.Lat != tt.lat || c.Lng != tt.lng) {
				t.Fatalf("got %v", c)
			}
			if c2 := (Coordinate{tt.lat, tt.lng}); c2.IsValid() != (tt.err == nil) {
				t.Fatalf("IsValid mismatch")
			}
		})
	}
}

func TestMustNewPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	MustNew(100, 0)
}

func TestMustNewOK(t *testing.T) {
	if c := MustNew(1, 2); c.Lat != 1 || c.Lng != 2 {
		t.Fatal(c)
	}
}

func TestEqual(t *testing.T) {
	a := Coordinate{1, 2}
	if !a.Equal(Coordinate{1.0000001, 2}, 1e-6) {
		t.Fatal("expected equal")
	}
	if a.Equal(Coordinate{1.1, 2}, 1e-6) {
		t.Fatal("expected not equal")
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct{ in, want Coordinate }{
		{Coordinate{10, 190}, Coordinate{10, -170}},
		{Coordinate{10, -190}, Coordinate{10, 170}},
		{Coordinate{10, 180}, Coordinate{10, -180}},
		{Coordinate{10, 540}, Coordinate{10, -180}},
		{Coordinate{95, 0}, Coordinate{90, 0}},
		{Coordinate{-95, 0}, Coordinate{-90, 0}},
		{Coordinate{41, 69}, Coordinate{41, 69}},
	}
	for _, tt := range tests {
		nearCoord(t, "Normalize", tt.in.Normalize(), tt.want, 1e-9)
	}
}

func TestStringZeroAntipode(t *testing.T) {
	if s := (Coordinate{41.3111, 69.2797}).String(); s != "41.311100,69.279700" {
		t.Fatal(s)
	}
	if !(Coordinate{}).IsZero() || tashkent.IsZero() {
		t.Fatal("IsZero")
	}
	nearCoord(t, "Antipode", tashkent.Antipode(), Coordinate{-41.2995, 69.2401 - 180}, 1e-9)
	nearCoord(t, "Antipode west", Coordinate{10, -30}.Antipode(), Coordinate{-10, 150}, 1e-9)
}
