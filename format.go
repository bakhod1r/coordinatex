package coordinatex

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// FormatDecimal returns "lat,lng" with prec decimal places.
func (c Coordinate) FormatDecimal(prec int) string {
	return strconv.FormatFloat(c.Lat, 'f', prec, 64) + "," + strconv.FormatFloat(c.Lng, 'f', prec, 64)
}

// FormatDMS returns degrees, minutes, seconds: 41°18'39.96"N 69°16'46.92"E.
func (c Coordinate) FormatDMS() string {
	return dms(c.Lat, "N", "S") + " " + dms(c.Lng, "E", "W")
}

// FormatDM returns degrees and decimal minutes: 41°18.666'N 69°16.782'E.
func (c Coordinate) FormatDM() string {
	return dm(c.Lat, "N", "S") + " " + dm(c.Lng, "E", "W")
}

func hemisphere(v float64, pos, neg string) string {
	if v < 0 {
		return neg
	}
	return pos
}

func dms(v float64, pos, neg string) string {
	total := int64(math.Round(math.Abs(v) * 360000)) // hundredths of arc-second
	d, rem := total/360000, total%360000
	m, cs := rem/6000, rem%6000
	return fmt.Sprintf("%d°%02d'%02d.%02d\"%s", d, m, cs/100, cs%100, hemisphere(v, pos, neg))
}

func dm(v float64, pos, neg string) string {
	total := int64(math.Round(math.Abs(v) * 60000)) // thousandths of arc-minute
	d, rem := total/60000, total%60000
	return fmt.Sprintf("%d°%02d.%03d'%s", d, rem/1000, rem%1000, hemisphere(v, pos, neg))
}

var (
	numberRe   = regexp.MustCompile(`^-?\d+(\.\d+)?`)
	symbolRepl = strings.NewReplacer("″", `"`, "“", `"`, "”", `"`, "′", "'", "’", "'", "º", "°", "''", `"`)
)

// Parse reads a coordinate pair in decimal ("41.3111,69.2797" or "41.3111 69.2797"),
// DMS (41°18'39.96"N 69°16'46.92"E), DM (41°18.666'N 69°16.782'E) or
// space-separated DMS (41 18 39.96 N 69 16 46.92 E). Hemisphere letters may be
// prefix or suffix. Latitude must come first.
func Parse(s string) (Coordinate, error) {
	s = strings.ToUpper(strings.TrimSpace(symbolRepl.Replace(s)))
	latS, lngS, ok := splitPair(s)
	if !ok {
		return Coordinate{}, fmt.Errorf("%w: %q", ErrParse, s)
	}
	lat, err := parseComponent(latS, "N", "S")
	if err != nil {
		return Coordinate{}, err
	}
	lng, err := parseComponent(lngS, "E", "W")
	if err != nil {
		return Coordinate{}, err
	}
	return New(lat, lng)
}

func splitPair(s string) (string, string, bool) {
	if strings.Count(s, ",") == 1 {
		a, b, _ := strings.Cut(s, ",")
		return a, b, true
	}
	if strings.Contains(s, ",") {
		return "", "", false
	}
	if i := strings.IndexAny(s, "NS"); i >= 0 {
		if i == 0 { // prefix form: N41 E69
			if j := strings.IndexAny(s, "EW"); j > 0 {
				return s[:j], s[j:], true
			}
			return "", "", false
		}
		return s[:i+1], s[i+1:], true
	}
	f := strings.Fields(s)
	if len(f) != 2 {
		return "", "", false
	}
	return f[0], f[1], true
}

func parseComponent(s, pos, neg string) (float64, error) {
	bad := func() (float64, error) { return 0, fmt.Errorf("%w: %q", ErrParse, s) }
	s = strings.TrimSpace(s)
	sign := 1.0
	hemi := ""
	for _, h := range []string{pos, neg, "N", "S", "E", "W"} {
		if strings.HasPrefix(s, h) || strings.HasSuffix(s, h) {
			hemi = h
			s = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(s, h), h))
			break
		}
	}
	if hemi != "" && hemi != pos && hemi != neg {
		return bad()
	}
	if hemi == neg {
		sign = -1
	}
	var parts []float64
	for s != "" {
		m := numberRe.FindString(s)
		if m == "" {
			return bad()
		}
		v, err := strconv.ParseFloat(m, 64)
		if err != nil {
			return bad()
		}
		if len(parts) > 0 && v < 0 {
			return bad()
		}
		parts = append(parts, v)
		s = strings.TrimLeft(s[len(m):], "°'\" \tD")
	}
	if len(parts) == 0 || len(parts) > 3 {
		return bad()
	}
	deg := parts[0]
	if deg < 0 {
		if hemi != "" {
			return bad()
		}
		sign, deg = -1, -deg
	}
	if len(parts) > 1 {
		if parts[1] >= 60 || (len(parts) == 3 && parts[1] != math.Trunc(parts[1])) {
			return bad()
		}
		deg += parts[1] / 60
	}
	if len(parts) == 3 {
		if parts[2] >= 60 {
			return bad()
		}
		deg += parts[2] / 3600
	}
	return sign * deg, nil
}
