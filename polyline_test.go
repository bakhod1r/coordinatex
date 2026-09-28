package coordinatex

import (
	"errors"
	"testing"
)

func TestPolylineLengthBounds(t *testing.T) {
	l := Polyline{tashkent, samarkand, tashkent}
	near(t, "Length", l.Length().Meters(), 2*DistanceBetween(tashkent, samarkand).Meters(), 1e-6)
	if (Polyline{}).Length() != 0 || (Polyline{tashkent}).Length() != 0 {
		t.Fatal("short lines")
	}
	b, err := l.Bounds()
	if err != nil || !b.Contains(tashkent) || !b.Contains(samarkand) {
		t.Fatal(b, err)
	}
}

func TestPolylineNearestPoint(t *testing.T) {
	l := Polyline{{0, 0}, {0, 10}, {10, 10}}
	p, d, ok := l.NearestPoint(Coordinate{1, 5})
	if !ok {
		t.Fatal("ok")
	}
	nearCoord(t, "projection", p, Coordinate{0, 5}, 1e-6)
	near(t, "dist", d.Kilometers(), 111.19, 0.05)
	p, _, _ = l.NearestPoint(Coordinate{-1, -1})
	nearCoord(t, "clamp to start", p, Coordinate{0, 0}, 1e-9)
	p, _, _ = l.NearestPoint(Coordinate{12, 10})
	nearCoord(t, "clamp to end", p, Coordinate{10, 10}, 1e-9)
	if _, _, ok := (Polyline{}).NearestPoint(tashkent); ok {
		t.Fatal("empty")
	}
	p, _, ok = (Polyline{tashkent}).NearestPoint(samarkand)
	if !ok || p != tashkent {
		t.Fatal("single")
	}
}

func TestPolylineSimplify(t *testing.T) {
	l := Polyline{{0, 0}, {0.00001, 0.5}, {0, 1}, {0.5, 1.00001}, {1, 1}}
	s := l.Simplify(100 * Meter)
	if len(s) != 3 || s[0] != l[0] || s[1] != l[2] || s[2] != l[4] {
		t.Fatal(s)
	}
	if got := l.Simplify(0); len(got) != len(l) {
		t.Fatal("zero tolerance keeps all", got)
	}
	if got := (Polyline{tashkent, samarkand}).Simplify(1e9); len(got) != 2 {
		t.Fatal(got)
	}
}

func TestEncodePolyline(t *testing.T) {
	l := Polyline{{38.5, -120.2}, {40.7, -120.95}, {43.252, -126.453}}
	const enc = "_p~iF~ps|U_ulLnnqC_mqNvxq`@"
	if s := EncodePolyline(l, 5); s != enc {
		t.Fatal(s)
	}
	got, err := DecodePolyline(enc, 5)
	if err != nil || len(got) != 3 {
		t.Fatal(got, err)
	}
	for i := range l {
		nearCoord(t, "decode", got[i], l[i], 1e-9)
	}
	// precision 6 round trip
	got, err = DecodePolyline(EncodePolyline(Polyline{tashkent, samarkand}, 6), 6)
	if err != nil {
		t.Fatal(err)
	}
	nearCoord(t, "p6", got[1], samarkand, 1e-6)
	for _, bad := range []string{"_p~iF", "_p~iF~ps|U_", "\x01\x02"} {
		if _, err := DecodePolyline(bad, 5); !errors.Is(err, ErrInvalidPolyline) {
			t.Errorf("%q: %v", bad, err)
		}
	}
	if got, err := DecodePolyline("", 5); err != nil || len(got) != 0 {
		t.Fatal("empty", got, err)
	}
}

func FuzzDecodePolyline(f *testing.F) {
	f.Add("_p~iF~ps|U_ulLnnqC_mqNvxq`@")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = DecodePolyline(s, 5)
	})
}

func FuzzDecodeGeohash(f *testing.F) {
	f.Add("u4pruydqqvj")
	f.Fuzz(func(t *testing.T, s string) {
		if c, err := DecodeGeohash(s); err == nil && !c.IsValid() {
			t.Fatalf("%q -> %v", s, c)
		}
	})
}
