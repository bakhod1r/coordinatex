package coordinatex

import (
	"fmt"
	"math"
	"strings"
)

// Web Mercator (EPSG:3857) sphere radius and latitude limit.
const (
	mercatorR      = 6378137.0
	MaxMercatorLat = 85.05112877980659
)

// ToWebMercator projects c to EPSG:3857 meters. Latitude is clamped to ±MaxMercatorLat.
func ToWebMercator(c Coordinate) (x, y float64) {
	lat := math.Max(-MaxMercatorLat, math.Min(MaxMercatorLat, c.Lat))
	return mercatorR * toRad(c.Lng), mercatorR * math.Log(math.Tan(math.Pi/4+toRad(lat)/2))
}

// FromWebMercator converts EPSG:3857 meters back to a coordinate.
func FromWebMercator(x, y float64) Coordinate {
	return Coordinate{Lat: toDeg(2*math.Atan(math.Exp(y/mercatorR)) - math.Pi/2), Lng: wrapLng(toDeg(x / mercatorR))}
}

// Tile is an XYZ (slippy map) tile.
type Tile struct{ X, Y, Z int }

// TileAt returns the tile containing c at zoom z.
func TileAt(c Coordinate, z int) Tile {
	n := 1 << z
	lat := toRad(math.Max(-MaxMercatorLat, math.Min(MaxMercatorLat, c.Lat)))
	x := int(math.Floor((c.Lng + 180) / 360 * float64(n)))
	y := int(math.Floor((1 - math.Asinh(math.Tan(lat))/math.Pi) / 2 * float64(n)))
	clamp := func(v int) int { return max(0, min(n-1, v)) }
	return Tile{X: clamp(x), Y: clamp(y), Z: z}
}

func tileLng(x, z int) float64 { return float64(x)/float64(int(1)<<z)*360 - 180 }
func tileLat(y, z int) float64 {
	return toDeg(math.Atan(math.Sinh(math.Pi * (1 - 2*float64(y)/float64(int(1)<<z)))))
}

// Bounds returns the tile's lat/lng extent.
func (t Tile) Bounds() Bounds {
	return Bounds{MinLat: tileLat(t.Y+1, t.Z), MinLng: tileLng(t.X, t.Z), MaxLat: tileLat(t.Y, t.Z), MaxLng: tileLng(t.X+1, t.Z)}
}

// Parent returns the enclosing tile one zoom level up.
func (t Tile) Parent() Tile { return Tile{X: t.X / 2, Y: t.Y / 2, Z: t.Z - 1} }

// Children returns the four tiles one zoom level down (NW, NE, SW, SE).
func (t Tile) Children() [4]Tile {
	x, y, z := t.X*2, t.Y*2, t.Z+1
	return [4]Tile{{x, y, z}, {x + 1, y, z}, {x, y + 1, z}, {x + 1, y + 1, z}}
}

// String formats as "z/x/y".
func (t Tile) String() string { return fmt.Sprintf("%d/%d/%d", t.Z, t.X, t.Y) }

// Quadkey returns the Bing Maps quadkey.
func (t Tile) Quadkey() string {
	var sb strings.Builder
	for i := t.Z; i > 0; i-- {
		d := byte('0')
		mask := 1 << (i - 1)
		if t.X&mask != 0 {
			d++
		}
		if t.Y&mask != 0 {
			d += 2
		}
		sb.WriteByte(d)
	}
	return sb.String()
}

// TileFromQuadkey parses a Bing Maps quadkey.
func TileFromQuadkey(q string) (Tile, error) {
	t := Tile{Z: len(q)}
	for i := 0; i < len(q); i++ {
		mask := 1 << (len(q) - 1 - i)
		switch q[i] {
		case '0':
		case '1':
			t.X |= mask
		case '2':
			t.Y |= mask
		case '3':
			t.X |= mask
			t.Y |= mask
		default:
			return Tile{}, fmt.Errorf("%w: quadkey %q", ErrParse, q)
		}
	}
	return t, nil
}

// TilesCovering returns every tile at zoom z intersecting b.
func TilesCovering(b Bounds, z int) []Tile {
	var out []Tile
	for _, iv := range b.lngIntervals() {
		nw := TileAt(Coordinate{Lat: b.MaxLat, Lng: iv[0]}, z)
		se := TileAt(Coordinate{Lat: b.MinLat, Lng: math.Nextafter(iv[1], iv[0])}, z)
		for x := nw.X; x <= se.X; x++ {
			for y := nw.Y; y <= se.Y; y++ {
				out = append(out, Tile{X: x, Y: y, Z: z})
			}
		}
	}
	return out
}
