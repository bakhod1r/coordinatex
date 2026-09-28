package coordinatex

import (
	"fmt"
	"strings"
)

const geohashAlphabet = "0123456789bcdefghjkmnpqrstuvwxyz"

// MaxGeohashPrecision is the longest geohash produced (~3.7 cm cells).
const MaxGeohashPrecision = 12

var geohashIndex = func() (t [256]int8) {
	for i := range t {
		t[i] = -1
	}
	for i := 0; i < len(geohashAlphabet); i++ {
		t[geohashAlphabet[i]] = int8(i)
		t[strings.ToUpper(geohashAlphabet[i : i+1])[0]] = int8(i)
	}
	return
}()

// EncodeGeohash encodes c with the given precision (1..12); out-of-range values use 12.
func EncodeGeohash(c Coordinate, precision int) string {
	if precision < 1 || precision > MaxGeohashPrecision {
		precision = MaxGeohashPrecision
	}
	latLo, latHi, lngLo, lngHi := -90.0, 90.0, -180.0, 180.0
	out := make([]byte, precision)
	even := true
	for i := range out {
		var idx byte
		for bit := 0; bit < 5; bit++ {
			idx <<= 1
			if even {
				if mid := (lngLo + lngHi) / 2; c.Lng >= mid {
					idx |= 1
					lngLo = mid
				} else {
					lngHi = mid
				}
			} else {
				if mid := (latLo + latHi) / 2; c.Lat >= mid {
					idx |= 1
					latLo = mid
				} else {
					latHi = mid
				}
			}
			even = !even
		}
		out[i] = geohashAlphabet[idx]
	}
	return string(out)
}

// GeohashBounds returns the cell covered by hash.
func GeohashBounds(hash string) (Bounds, error) {
	if hash == "" || len(hash) > MaxGeohashPrecision {
		return Bounds{}, fmt.Errorf("%w: %q", ErrInvalidGeohash, hash)
	}
	b := Bounds{MinLat: -90, MaxLat: 90, MinLng: -180, MaxLng: 180}
	even := true
	for i := 0; i < len(hash); i++ {
		idx := geohashIndex[hash[i]]
		if idx < 0 {
			return Bounds{}, fmt.Errorf("%w: %q", ErrInvalidGeohash, hash)
		}
		for bit := 4; bit >= 0; bit-- {
			on := idx>>bit&1 == 1
			if even {
				mid := (b.MinLng + b.MaxLng) / 2
				if on {
					b.MinLng = mid
				} else {
					b.MaxLng = mid
				}
			} else {
				mid := (b.MinLat + b.MaxLat) / 2
				if on {
					b.MinLat = mid
				} else {
					b.MaxLat = mid
				}
			}
			even = !even
		}
	}
	return b, nil
}

// DecodeGeohash returns the center of the hash cell. Decoding is case-insensitive.
func DecodeGeohash(hash string) (Coordinate, error) {
	b, err := GeohashBounds(hash)
	if err != nil {
		return Coordinate{}, err
	}
	return b.Center(), nil
}

// GeohashNeighbors returns the 8 adjacent cells in order N, NE, E, SE, S, SW, W, NW.
// Longitude wraps across the antimeridian; at the poles the northern/southern
// neighbours are clamped to the edge cell.
func GeohashNeighbors(hash string) ([8]string, error) {
	var out [8]string
	b, err := GeohashBounds(hash)
	if err != nil {
		return out, err
	}
	c := b.Center()
	dLat, dLng := b.MaxLat-b.MinLat, b.MaxLng-b.MinLng
	offsets := [8][2]float64{{1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}, {1, -1}}
	for i, o := range offsets {
		lat := c.Lat + o[0]*dLat
		if lat > 90 || lat < -90 {
			lat = c.Lat
		}
		out[i] = EncodeGeohash(Coordinate{Lat: lat, Lng: wrapLng(c.Lng + o[1]*dLng)}, len(hash))
	}
	return out, nil
}

// geohashCellWidth is the approximate cell width at the equator for each precision.
var geohashCellWidth = [MaxGeohashPrecision + 1]Distance{
	0, 5009.4 * Kilometer, 1252.3 * Kilometer, 156.5 * Kilometer, 39.1 * Kilometer,
	4.89 * Kilometer, 1.22 * Kilometer, 152.9, 38.2, 4.77, 1.19, 0.149, 0.037,
}

// GeohashPrecisionFor returns the highest precision whose cell width is at least d,
// so a hash plus its 8 neighbours covers a search radius of d.
func GeohashPrecisionFor(d Distance) int {
	for p := MaxGeohashPrecision; p >= 1; p-- {
		if geohashCellWidth[p] >= d {
			return p
		}
	}
	return 1
}

// IsValidGeohash reports whether hash is a well-formed geohash.
func IsValidGeohash(hash string) bool {
	_, err := GeohashBounds(hash)
	return err == nil
}

// GeohashParent returns hash without its last character.
func GeohashParent(hash string) (string, error) {
	if !IsValidGeohash(hash) || len(hash) < 2 {
		return "", fmt.Errorf("%w: no parent for %q", ErrInvalidGeohash, hash)
	}
	return hash[:len(hash)-1], nil
}

// GeohashChildren returns the 32 cells one level deeper.
func GeohashChildren(hash string) ([]string, error) {
	if !IsValidGeohash(hash) || len(hash) >= MaxGeohashPrecision {
		return nil, fmt.Errorf("%w: no children for %q", ErrInvalidGeohash, hash)
	}
	out := make([]string, len(geohashAlphabet))
	for i := range out {
		out[i] = hash + geohashAlphabet[i:i+1]
	}
	return out, nil
}

// GeohashesAround returns the cell containing center plus its 8 neighbours at
// the finest precision whose cells (at that latitude) are at least radius wide
// and tall, so every point within radius falls in one of them. Use as the
// candidate filter of a two-stage search (prefix match, then WithinRadius).
func GeohashesAround(center Coordinate, radius Distance) []string {
	p := 1
	for q := MaxGeohashPrecision; q >= 1; q-- {
		b, _ := GeohashBounds(EncodeGeohash(center, q))
		if b.Width() >= radius && b.Height() >= radius {
			p = q
			break
		}
	}
	h := EncodeGeohash(center, p)
	n, _ := GeohashNeighbors(h)
	return append([]string{h}, n[:]...)
}
