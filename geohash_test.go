package coordinatex

import (
	"errors"
	"testing"
)

func TestEncodeGeohash(t *testing.T) {
	c := Coordinate{57.64911, 10.40744}
	if h := EncodeGeohash(c, 11); h != "u4pruydqqvj" {
		t.Fatal(h)
	}
	if h := EncodeGeohash(c, 5); h != "u4pru" {
		t.Fatal(h)
	}
	if h := EncodeGeohash(c, 0); len(h) != 12 {
		t.Fatal("default precision", h)
	}
	if h := EncodeGeohash(c, 40); len(h) != 12 {
		t.Fatal("max precision", h)
	}
}

func TestDecodeGeohash(t *testing.T) {
	c, err := DecodeGeohash("u4pruydqqvj")
	if err != nil {
		t.Fatal(err)
	}
	nearCoord(t, "decode", c, Coordinate{57.64911, 10.40744}, 1e-5)
	c, err = DecodeGeohash("U4PRU") // case-insensitive
	if err != nil || !c.Equal(Coordinate{57.65, 10.41}, 0.03) {
		t.Fatal(c, err)
	}
	for _, bad := range []string{"", "u4pa", "abc!", "0123456789bcdefghj"} {
		if _, err := DecodeGeohash(bad); !errors.Is(err, ErrInvalidGeohash) {
			t.Errorf("%q: %v", bad, err)
		}
	}
}

func TestGeohashRoundTrip(t *testing.T) {
	for _, c := range []Coordinate{tashkent, samarkand, london, paris, {-33.8688, 151.2093}, {0, 0}, {89.9, 179.9}, {-89.9, -179.9}} {
		for p := 1; p <= 12; p++ {
			h := EncodeGeohash(c, p)
			b, err := GeohashBounds(h)
			if err != nil || !b.Contains(c) {
				t.Fatalf("%v p%d %s %+v %v", c, p, h, b, err)
			}
		}
	}
}

func TestGeohashNeighbors(t *testing.T) {
	n, err := GeohashNeighbors("gbsuv")
	if err != nil {
		t.Fatal(err)
	}
	want := [8]string{"gbsvj", "gbsvn", "gbsuy", "gbsuw", "gbsut", "gbsus", "gbsuu", "gbsvh"}
	if n != want {
		t.Fatalf("got %v want %v", n, want)
	}
	// wraps across antimeridian
	e, err := GeohashNeighbors(EncodeGeohash(Coordinate{0.1, 179.99}, 5))
	if err != nil {
		t.Fatal(err)
	}
	c, _ := DecodeGeohash(e[2])
	if c.Lng > 0 {
		t.Fatalf("east neighbor should wrap: %v", c)
	}
	if _, err := GeohashNeighbors("!!"); err == nil {
		t.Fatal("invalid accepted")
	}
}

func TestGeohashPrecisionFor(t *testing.T) {
	cases := map[Distance]int{5 * Kilometer: 4, 1 * Kilometer: 6, 100 * Meter: 7, 1 * Meter: 10, 1 * 1e-3: 12, 10000 * Kilometer: 1}
	for d, want := range cases {
		if got := GeohashPrecisionFor(d); got != want {
			t.Errorf("%v: got %d want %d", d, got, want)
		}
	}
}

func BenchmarkEncodeGeohash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeGeohash(tashkent, 9)
	}
}
