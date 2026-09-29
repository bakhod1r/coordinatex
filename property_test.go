package coordinatex

import (
	"math"
	"math/rand"
	"testing"
	"testing/quick"
)

func randCoord(r *rand.Rand) Coordinate {
	return Coordinate{Lat: r.Float64()*170 - 85, Lng: r.Float64()*360 - 180}
}

func quickCheck(t *testing.T, f func(r *rand.Rand) bool) {
	t.Helper()
	r := rand.New(rand.NewSource(42))
	if err := quick.Check(func(seed int64) bool { r.Seed(seed); return f(r) }, &quick.Config{MaxCount: 2000}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyDestinationInverse(t *testing.T) {
	quickCheck(t, func(r *rand.Rand) bool {
		a := randCoord(r)
		d := Distance(r.Float64() * 5000 * float64(Kilometer))
		b := Destination(a, d, Angle(r.Float64()*360))
		return math.Abs(DistanceBetween(a, b).Meters()-d.Meters()) < 1e-3
	})
}

func TestPropertyBearingRoundTrip(t *testing.T) {
	quickCheck(t, func(r *rand.Rand) bool {
		a, b := randCoord(r), randCoord(r)
		d := DistanceBetween(a, b)
		if d < 1 || d > 19000*Kilometer { // skip (near-)antipodes where bearing is unstable
			return true
		}
		return DistanceBetween(Destination(a, d, Bearing(a, b)), b) < 1
	})
}

func TestPropertyTriangleInequality(t *testing.T) {
	quickCheck(t, func(r *rand.Rand) bool {
		a, b, c := randCoord(r), randCoord(r), randCoord(r)
		return DistanceBetween(a, c) <= DistanceBetween(a, b)+DistanceBetween(b, c)+1e-6
	})
}

func TestPropertyGeohashContains(t *testing.T) {
	quickCheck(t, func(r *rand.Rand) bool {
		c := randCoord(r)
		b, err := GeohashBounds(EncodeGeohash(c, 1+r.Intn(12)))
		return err == nil && b.Contains(c)
	})
}

func TestPropertyVincentyCloseToHaversine(t *testing.T) {
	quickCheck(t, func(r *rand.Rand) bool {
		a, b := randCoord(r), randCoord(r)
		v, err := DistanceVincenty(a, b)
		if err != nil {
			return true
		}
		h := DistanceBetween(a, b)
		return math.Abs(v.Meters()-h.Meters()) <= 0.006*h.Meters()+1 // sphere vs ellipsoid ≤ ~0.5%
	})
}

func TestPropertyVincentyDirectInverse(t *testing.T) {
	quickCheck(t, func(r *rand.Rand) bool {
		a := randCoord(r)
		d := Distance(r.Float64() * 10000 * float64(Kilometer))
		b, err := DestinationVincenty(a, d, Angle(r.Float64()*360))
		if err != nil {
			return false
		}
		back, err := DistanceVincenty(a, b)
		return err != nil || math.Abs(back.Meters()-d.Meters()) < 1e-3
	})
}

func TestPropertyUTMRoundTrip(t *testing.T) {
	quickCheck(t, func(r *rand.Rand) bool {
		c := Coordinate{Lat: r.Float64()*163 - 79.5, Lng: r.Float64()*360 - 180}
		u, err := ToUTM(c)
		if err != nil {
			return false
		}
		back, err := u.Coordinate()
		return err == nil && back.Equal(c, 1e-7)
	})
}
