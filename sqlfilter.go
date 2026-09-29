package coordinatex

import (
	"fmt"
	"regexp"
	"strings"
)

// Dialect selects the SQL placeholder style.
type Dialect int

const (
	// Postgres uses numbered placeholders ($1, $2, ...).
	Postgres Dialect = iota
	// MySQL uses ? placeholders (also SQLite).
	MySQL
)

// SQLFilter is a parameterized WHERE fragment. Values are always passed as
// arguments; column names are validated as plain identifiers.
type SQLFilter struct {
	Clause string
	Args   []any
}

// And joins two filters. b must have been built with argOffset = len(f.Args).
func (f SQLFilter) And(b SQLFilter) SQLFilter {
	return SQLFilter{Clause: f.Clause + " AND " + b.Clause, Args: append(append([]any{}, f.Args...), b.Args...)}
}

var identifierRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)?$`)

func checkIdent(cols ...string) error {
	for _, c := range cols {
		if !identifierRe.MatchString(c) {
			return fmt.Errorf("%w: %q", ErrInvalidIdentifier, c)
		}
	}
	return nil
}

// builder accumulates placeholders and arguments.
type builder struct {
	d    Dialect
	n    int
	args []any
}

func (d Dialect) builder(argOffset int) (*builder, error) {
	if argOffset < 0 {
		return nil, fmt.Errorf("coordinatex: negative argOffset %d", argOffset)
	}
	return &builder{d: d, n: argOffset}, nil
}

func (b *builder) arg(v any) string {
	b.n++
	b.args = append(b.args, v)
	if b.d == MySQL {
		return "?"
	}
	return fmt.Sprintf("$%d", b.n)
}

// BoundsFilter matches rows whose latitude/longitude columns fall inside b,
// splitting the longitude test when b crosses the antimeridian. argOffset is
// the number of arguments already used by the surrounding query.
func (d Dialect) BoundsFilter(latCol, lngCol string, b Bounds, argOffset int) (SQLFilter, error) {
	if err := checkIdent(latCol, lngCol); err != nil {
		return SQLFilter{}, err
	}
	q, err := d.builder(argOffset)
	if err != nil {
		return SQLFilter{}, err
	}
	lat := fmt.Sprintf("%s BETWEEN %s AND %s", latCol, q.arg(b.MinLat), q.arg(b.MaxLat))
	var lng string
	if b.CrossesAntimeridian() {
		lng = fmt.Sprintf("(%s >= %s OR %s <= %s)", lngCol, q.arg(b.MinLng), lngCol, q.arg(b.MaxLng))
	} else {
		lng = fmt.Sprintf("%s BETWEEN %s AND %s", lngCol, q.arg(b.MinLng), q.arg(b.MaxLng))
	}
	return SQLFilter{Clause: "(" + lat + " AND " + lng + ")", Args: q.args}, nil
}

// RadiusFilter is the index-friendly first stage of a radius search:
// BoundsFilter over BoundsAround(center, r). Confirm hits with WithinRadius.
func (d Dialect) RadiusFilter(latCol, lngCol string, center Coordinate, r Distance, argOffset int) (SQLFilter, error) {
	return d.BoundsFilter(latCol, lngCol, BoundsAround(center, r), argOffset)
}

func (b *builder) point(c Coordinate) string {
	return fmt.Sprintf("ST_SetSRID(ST_MakePoint(%s, %s), 4326)", b.arg(c.Lng), b.arg(c.Lat))
}

// DWithin is the PostGIS geography radius test ST_DWithin(col, point, meters).
// It uses a GiST index on the geography column.
func (d Dialect) DWithin(geogCol string, center Coordinate, r Distance, argOffset int) (SQLFilter, error) {
	if err := checkIdent(geogCol); err != nil {
		return SQLFilter{}, err
	}
	if err := center.Validate(); err != nil {
		return SQLFilter{}, err
	}
	q, err := d.builder(argOffset)
	if err != nil {
		return SQLFilter{}, err
	}
	clause := fmt.Sprintf("ST_DWithin(%s, %s::geography, %s)", geogCol, q.point(center), q.arg(r.Meters()))
	return SQLFilter{Clause: clause, Args: q.args}, nil
}

// DistanceOrder is a KNN ORDER BY expression (col <-> point) for nearest-first results.
func (d Dialect) DistanceOrder(geogCol string, center Coordinate, argOffset int) (SQLFilter, error) {
	if err := checkIdent(geogCol); err != nil {
		return SQLFilter{}, err
	}
	q, err := d.builder(argOffset)
	if err != nil {
		return SQLFilter{}, err
	}
	return SQLFilter{Clause: fmt.Sprintf("%s <-> %s::geography", geogCol, q.point(center)), Args: q.args}, nil
}

// Envelope matches geometries intersecting b's bounding box (geom && ST_MakeEnvelope),
// splitting into two envelopes when b crosses the antimeridian.
func (d Dialect) Envelope(geomCol string, b Bounds, argOffset int) (SQLFilter, error) {
	if err := checkIdent(geomCol); err != nil {
		return SQLFilter{}, err
	}
	q, err := d.builder(argOffset)
	if err != nil {
		return SQLFilter{}, err
	}
	env := func(minLng, maxLng float64) string {
		return fmt.Sprintf("%s && ST_MakeEnvelope(%s, %s, %s, %s, 4326)", geomCol, q.arg(minLng), q.arg(b.MinLat), q.arg(maxLng), q.arg(b.MaxLat))
	}
	if b.CrossesAntimeridian() {
		return SQLFilter{Clause: "(" + env(b.MinLng, 180) + " OR " + env(-180, b.MaxLng) + ")", Args: q.args}, nil
	}
	return SQLFilter{Clause: env(b.MinLng, b.MaxLng), Args: q.args}, nil
}

// GeohashPrefixFilter matches rows whose geohash column starts with any of the
// cells (for example the output of GeohashesAround).
func (d Dialect) GeohashPrefixFilter(col string, cells []string, argOffset int) (SQLFilter, error) {
	if err := checkIdent(col); err != nil {
		return SQLFilter{}, err
	}
	if len(cells) == 0 {
		return SQLFilter{}, ErrEmpty
	}
	q, err := d.builder(argOffset)
	if err != nil {
		return SQLFilter{}, err
	}
	parts := make([]string, len(cells))
	for i, c := range cells {
		if !IsValidGeohash(c) {
			return SQLFilter{}, fmt.Errorf("%w: %q", ErrInvalidGeohash, c)
		}
		parts[i] = fmt.Sprintf("%s LIKE %s", col, q.arg(strings.ToLower(c)+"%"))
	}
	return SQLFilter{Clause: "(" + strings.Join(parts, " OR ") + ")", Args: q.args}, nil
}
