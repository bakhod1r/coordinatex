package coordinatex

import (
	"fmt"
	"math"
	"strings"
)

const (
	plusAlphabet  = "23456789CFGHJMPQRVWX"
	plusSeparator = 8
	plusLatPrec   = 25000000 // 8000 * 5^5
	plusLngPrec   = 8192000  // 8000 * 4^5
)

var plusPlaceValues = [5]float64{20, 1, 0.05, 0.0025, 0.000125}

// EncodePlusCode returns the Open Location Code (Plus Code) for c.
// length is 2, 4, 6, 8, 10 (≈14 m) or 11..15 (grid refinement).
func EncodePlusCode(c Coordinate, length int) string {
	switch {
	case length < 2:
		length = 2
	case length < 10 && length%2 == 1:
		length++
	case length > 15:
		length = 15
	}
	lat := math.Max(-90, math.Min(90, c.Lat))
	latVal := int64(math.Floor(math.Round((lat+90)*plusLatPrec*1e6) / 1e6))
	lngVal := int64(math.Floor(math.Round((wrapLng(c.Lng)+180)*plusLngPrec*1e6) / 1e6))
	latVal = min(latVal, 180*plusLatPrec-1)
	lngVal %= 360 * plusLngPrec

	code := make([]byte, 15)
	for i := 14; i >= 10; i-- {
		code[i] = plusAlphabet[(latVal%5)*4+lngVal%4]
		latVal /= 5
		lngVal /= 4
	}
	for i := 8; i >= 0; i -= 2 {
		code[i+1] = plusAlphabet[lngVal%20]
		code[i] = plusAlphabet[latVal%20]
		latVal /= 20
		lngVal /= 20
	}
	digits := string(code[:length])
	if length < plusSeparator {
		digits += strings.Repeat("0", plusSeparator-length)
	}
	return digits[:plusSeparator] + "+" + digits[plusSeparator:]
}

// IsValidPlusCode reports whether s is a valid full Plus Code.
func IsValidPlusCode(s string) bool {
	_, err := plusDigits(s)
	return err == nil
}

func plusDigits(s string) (string, error) {
	bad := fmt.Errorf("%w: plus code %q", ErrParse, s)
	s = strings.ToUpper(s)
	if strings.IndexByte(s, '+') != plusSeparator || strings.Count(s, "+") != 1 {
		return "", bad
	}
	head, tail := s[:plusSeparator], s[plusSeparator+1:]
	if len(tail) == 1 {
		return "", bad
	}
	if pad := strings.IndexByte(head, '0'); pad >= 0 {
		if pad%2 == 1 || pad == 0 || strings.Trim(head[pad:], "0") != "" || tail != "" {
			return "", bad
		}
		head = head[:pad]
	}
	d := head + tail
	for i := 0; i < len(d); i++ {
		if strings.IndexByte(plusAlphabet, d[i]) < 0 {
			return "", bad
		}
	}
	if strings.IndexByte(plusAlphabet, d[0]) >= 9 || strings.IndexByte(plusAlphabet, d[1]) >= 18 {
		return "", bad
	}
	return d, nil
}

// DecodePlusCode returns the area covered by a full Plus Code.
func DecodePlusCode(s string) (Bounds, error) {
	d, err := plusDigits(s)
	if err != nil {
		return Bounds{}, err
	}
	lat, lng := -90.0, -180.0
	size := 0.0
	for i := 0; i < len(d) && i < 10; i += 2 {
		size = plusPlaceValues[i/2]
		lat += float64(strings.IndexByte(plusAlphabet, d[i])) * size
		lng += float64(strings.IndexByte(plusAlphabet, d[i+1])) * size
	}
	latSize, lngSize := size, size
	for i := 10; i < len(d); i++ {
		idx := strings.IndexByte(plusAlphabet, d[i])
		latSize /= 5
		lngSize /= 4
		lat += float64(idx/4) * latSize
		lng += float64(idx%4) * lngSize
	}
	return Bounds{MinLat: lat, MinLng: lng, MaxLat: math.Min(90, lat+latSize), MaxLng: lng + lngSize}, nil
}
