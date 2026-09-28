package coordinatex

import (
	"errors"
	"testing"
)

func TestFormat(t *testing.T) {
	c := Coordinate{41.3111, 69.2797}
	if s := c.FormatDMS(); s != `41°18'39.96"N 69°16'46.92"E` {
		t.Fatal(s)
	}
	if s := c.FormatDM(); s != `41°18.666'N 69°16.782'E` {
		t.Fatal(s)
	}
	if s := (Coordinate{-33.8688, -70.5}).FormatDMS(); s != `33°52'07.68"S 70°30'00.00"W` {
		t.Fatal(s)
	}
	if s := c.FormatDecimal(2); s != "41.31,69.28" {
		t.Fatal(s)
	}
	// rounding carry: 59.999" must become next minute
	if s := (Coordinate{10.9999999, 0}).FormatDMS(); s != `11°00'00.00"N 0°00'00.00"E` {
		t.Fatal(s)
	}
}

func TestParse(t *testing.T) {
	want := Coordinate{41.3111, 69.2797}
	for _, in := range []string{
		"41.3111,69.2797",
		" 41.3111 , 69.2797 ",
		"41.3111 69.2797",
		`41°18'39.96"N 69°16'46.92"E`,
		`41°18'39.96"N, 69°16'46.92"E`,
		`N41°18'39.96" E69°16'46.92"`,
		`41°18′39.96″N 69°16′46.92″E`,
		`41 18 39.96 N 69 16 46.92 E`,
		`41°18.666'N 69°16.782'E`,
		`41.3111N 69.2797E`,
	} {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		nearCoord(t, in, got, want, 1e-6)
	}
	got, err := Parse(`33°52'07.68"S 70°30'00"W`)
	if err != nil {
		t.Fatal(err)
	}
	nearCoord(t, "south-west", got, Coordinate{-33.8688, -70.5}, 1e-6)
	got, err = Parse("-33.8688,-70.5")
	if err != nil {
		t.Fatal(err)
	}
	nearCoord(t, "negative", got, Coordinate{-33.8688, -70.5}, 1e-9)
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{
		"", "abc", "41.3", "1,2,3", "91,0", "0,181",
		`41°61'00"N 69°E`, `41°10'61"N 69°E`, `69°E 41°N`,
		"41.3x,69", "N41 N69", "--1,2",
	} {
		if _, err := Parse(in); err == nil {
			t.Errorf("%q: expected error", in)
		} else if !errors.Is(err, ErrParse) && !errors.Is(err, ErrInvalidLatitude) && !errors.Is(err, ErrInvalidLongitude) {
			t.Errorf("%q: unexpected error type %v", in, err)
		}
	}
}

func TestParseFormatRoundTrip(t *testing.T) {
	for _, c := range []Coordinate{tashkent, samarkand, london, paris, {-33.8688, 151.2093}} {
		got, err := Parse(c.FormatDMS())
		if err != nil {
			t.Fatal(err)
		}
		nearCoord(t, "DMS round trip", got, c, 1e-5)
		got, err = Parse(c.FormatDM())
		if err != nil {
			t.Fatal(err)
		}
		nearCoord(t, "DM round trip", got, c, 1e-5)
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"41.3,69.2", `41°18'39.96"N 69°16'46.92"E`, "N41 E69", "-1 -2"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		c, err := Parse(s)
		if err == nil && !c.IsValid() {
			t.Fatalf("Parse(%q) returned invalid %v", s, c)
		}
	})
}
