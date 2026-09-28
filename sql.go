package coordinatex

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// WKT returns the OGC Well-Known Text form "POINT(lng lat)".
func (c Coordinate) WKT() string {
	return "POINT(" + strconv.FormatFloat(c.Lng, 'f', -1, 64) + " " + strconv.FormatFloat(c.Lat, 'f', -1, 64) + ")"
}

// ParseWKT parses "POINT(lng lat)", optionally prefixed with "SRID=n;" (EWKT).
func ParseWKT(s string) (Coordinate, error) {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ";"); i >= 0 && strings.HasPrefix(strings.ToUpper(s), "SRID=") {
		s = strings.TrimSpace(s[i+1:])
	}
	up := strings.ToUpper(s)
	if !strings.HasPrefix(up, "POINT") || !strings.HasSuffix(s, ")") {
		return Coordinate{}, fmt.Errorf("%w: WKT %q", ErrParse, s)
	}
	body := strings.TrimSpace(s[len("POINT"):])
	if !strings.HasPrefix(body, "(") {
		return Coordinate{}, fmt.Errorf("%w: WKT %q", ErrParse, s)
	}
	f := strings.Fields(body[1 : len(body)-1])
	if len(f) != 2 {
		return Coordinate{}, fmt.Errorf("%w: WKT %q", ErrParse, s)
	}
	lng, err1 := strconv.ParseFloat(f[0], 64)
	lat, err2 := strconv.ParseFloat(f[1], 64)
	if err1 != nil || err2 != nil {
		return Coordinate{}, fmt.Errorf("%w: WKT %q", ErrParse, s)
	}
	return New(lat, lng)
}

// Value implements driver.Valuer, storing the coordinate as WKT. With PostGIS use
// ST_GeomFromText($1, 4326) or a geography column that accepts WKT input.
func (c Coordinate) Value() (driver.Value, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c.WKT(), nil
}

// Scan implements sql.Scanner for WKT/EWKT text (SELECT ST_AsText(geom)),
// hex (E)WKB as returned by PostGIS for a plain geometry column, and binary WKB.
func (c *Coordinate) Scan(src any) error {
	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		if len(v) > 0 && v[0] <= 1 {
			p, err := ParseWKB(v)
			if err != nil {
				return err
			}
			*c = p
			return nil
		}
		s = string(v)
	default:
		return fmt.Errorf("%w: cannot scan %T", ErrParse, src)
	}
	if b, err := hex.DecodeString(s); err == nil && len(b) > 0 {
		p, err := ParseWKB(b)
		if err != nil {
			return err
		}
		*c = p
		return nil
	}
	p, err := ParseWKT(s)
	if err != nil {
		return err
	}
	*c = p
	return nil
}

const (
	wkbPoint     = 1
	ewkbSRID     = 0x20000000
	ewkbZ, ewkbM = 0x80000000, 0x40000000
)

// ParseWKB decodes a 2D Point in WKB or PostGIS EWKB (with optional SRID).
func ParseWKB(b []byte) (Coordinate, error) {
	if len(b) < 5 || b[0] > 1 {
		return Coordinate{}, fmt.Errorf("%w: WKB header", ErrParse)
	}
	var order binary.ByteOrder = binary.BigEndian
	if b[0] == 1 {
		order = binary.LittleEndian
	}
	typ := order.Uint32(b[1:5])
	b = b[5:]
	if typ&(ewkbZ|ewkbM) != 0 || typ&0xffff != wkbPoint {
		return Coordinate{}, fmt.Errorf("%w: WKB type %#x is not a 2D Point", ErrParse, typ)
	}
	if typ&ewkbSRID != 0 {
		if len(b) < 4 {
			return Coordinate{}, fmt.Errorf("%w: WKB SRID", ErrParse)
		}
		b = b[4:]
	}
	if len(b) != 16 {
		return Coordinate{}, fmt.Errorf("%w: WKB point length %d", ErrParse, len(b))
	}
	lng := math.Float64frombits(order.Uint64(b[:8]))
	lat := math.Float64frombits(order.Uint64(b[8:]))
	return New(lat, lng)
}

// EWKBHex encodes c as little-endian hex EWKB. srid <= 0 omits the SRID (plain WKB).
func (c Coordinate) EWKBHex(srid int) string {
	b := []byte{1}
	typ := uint32(wkbPoint)
	if srid > 0 {
		typ |= ewkbSRID
	}
	b = binary.LittleEndian.AppendUint32(b, typ)
	if srid > 0 {
		b = binary.LittleEndian.AppendUint32(b, uint32(srid))
	}
	b = binary.LittleEndian.AppendUint64(b, math.Float64bits(c.Lng))
	b = binary.LittleEndian.AppendUint64(b, math.Float64bits(c.Lat))
	return strings.ToUpper(hex.EncodeToString(b))
}
