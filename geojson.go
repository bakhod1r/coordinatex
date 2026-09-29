package coordinatex

import (
	"encoding/json"
	"fmt"
)

// Geometry is a value that can be encoded as a GeoJSON geometry (RFC 7946).
// Implemented by Coordinate (Point), MultiPoint, Polyline (LineString),
// MultiLineString, Polygon and PolygonWithHoles (Polygon), MultiPolygon and
// GeometryCollection.
type Geometry interface {
	GeoJSONType() string
	geoJSONCoordinates() (any, error)
}

// MultiPoint is a GeoJSON MultiPoint.
type MultiPoint []Coordinate

// MultiLineString is a GeoJSON MultiLineString.
type MultiLineString []Polyline

// MultiPolygon is a GeoJSON MultiPolygon.
type MultiPolygon []PolygonWithHoles

// GeometryCollection is a GeoJSON GeometryCollection.
type GeometryCollection []Geometry

func (Coordinate) GeoJSONType() string         { return "Point" }
func (MultiPoint) GeoJSONType() string         { return "MultiPoint" }
func (Polyline) GeoJSONType() string           { return "LineString" }
func (MultiLineString) GeoJSONType() string    { return "MultiLineString" }
func (Polygon) GeoJSONType() string            { return "Polygon" }
func (PolygonWithHoles) GeoJSONType() string   { return "Polygon" }
func (MultiPolygon) GeoJSONType() string       { return "MultiPolygon" }
func (GeometryCollection) GeoJSONType() string { return "GeometryCollection" }

type position [2]float64 // [lng, lat]

func toPosition(c Coordinate) position { return position{c.Lng, c.Lat} }

func positions(cs []Coordinate) ([]position, error) {
	out := make([]position, len(cs))
	for i, c := range cs {
		if err := c.Validate(); err != nil {
			return nil, err
		}
		out[i] = toPosition(c)
	}
	return out, nil
}

// ringPositions closes the ring and winds it counter-clockwise (exterior) or
// clockwise (hole), keeping the original first vertex first.
func ringPositions(p Polygon, ccw bool) ([]position, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	r := append([]Coordinate{}, p.ring()...)
	if p.IsClockwise() == ccw {
		for i, j := 1, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
	}
	return positions(append(r, r[0]))
}

func (c Coordinate) geoJSONCoordinates() (any, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return toPosition(c), nil
}

func (m MultiPoint) geoJSONCoordinates() (any, error) { return positions(m) }

func (l Polyline) geoJSONCoordinates() (any, error) { return positions(l) }

func (m MultiLineString) geoJSONCoordinates() (any, error) {
	out := make([][]position, len(m))
	for i, l := range m {
		p, err := positions(l)
		if err != nil {
			return nil, err
		}
		out[i] = p
	}
	return out, nil
}

func (p Polygon) geoJSONCoordinates() (any, error) {
	return PolygonWithHoles{Outer: p}.geoJSONCoordinates()
}

func (p PolygonWithHoles) geoJSONCoordinates() (any, error) {
	outer, err := ringPositions(p.Outer, true)
	if err != nil {
		return nil, err
	}
	rings := [][]position{outer}
	for _, h := range p.Holes {
		r, err := ringPositions(h, false)
		if err != nil {
			return nil, err
		}
		rings = append(rings, r)
	}
	return rings, nil
}

func (m MultiPolygon) geoJSONCoordinates() (any, error) {
	out := make([]any, len(m))
	for i, p := range m {
		c, err := p.geoJSONCoordinates()
		if err != nil {
			return nil, err
		}
		out[i] = c
	}
	return out, nil
}

// geoJSONCoordinates returns the encoded member geometries.
func (g GeometryCollection) geoJSONCoordinates() (any, error) {
	out := make([]json.RawMessage, len(g))
	for i, sub := range g {
		b, err := MarshalGeometry(sub)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

type rawGeometry struct {
	Type        string            `json:"type"`
	Coordinates json.RawMessage   `json:"coordinates,omitempty"`
	Geometries  []json.RawMessage `json:"geometries,omitempty"`
}

// MarshalGeometry encodes g as a GeoJSON geometry object.
func MarshalGeometry(g Geometry) ([]byte, error) {
	if g == nil {
		return nil, fmt.Errorf("%w: nil geometry", ErrParse)
	}
	coords, err := g.geoJSONCoordinates()
	if err != nil {
		return nil, err
	}
	raw := rawGeometry{Type: g.GeoJSONType()}
	if members, ok := coords.([]json.RawMessage); ok {
		raw.Geometries = members
	} else {
		// Positions are validated finite numbers, so encoding cannot fail.
		raw.Coordinates, _ = json.Marshal(coords)
	}
	return json.Marshal(raw)
}

// ParseGeometry decodes any GeoJSON geometry. Polygons decode as PolygonWithHoles.
func ParseGeometry(data []byte) (Geometry, error) {
	var raw rawGeometry
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}
	if raw.Type == "GeometryCollection" {
		gc := make(GeometryCollection, len(raw.Geometries))
		for i, sub := range raw.Geometries {
			g, err := ParseGeometry(sub)
			if err != nil {
				return nil, err
			}
			gc[i] = g
		}
		return gc, nil
	}
	decode := func(v any) error {
		if err := json.Unmarshal(raw.Coordinates, v); err != nil {
			return fmt.Errorf("%w: %s coordinates: %v", ErrParse, raw.Type, err)
		}
		return nil
	}
	switch raw.Type {
	case "Point":
		var p []float64
		if err := decode(&p); err != nil {
			return nil, err
		}
		return pointFrom(p)
	case "MultiPoint":
		var ps [][]float64
		if err := decode(&ps); err != nil {
			return nil, err
		}
		cs, err := coordsFrom(ps)
		return MultiPoint(cs), err
	case "LineString":
		var ps [][]float64
		if err := decode(&ps); err != nil {
			return nil, err
		}
		return lineFrom(ps)
	case "MultiLineString":
		var ls [][][]float64
		if err := decode(&ls); err != nil {
			return nil, err
		}
		out := make(MultiLineString, len(ls))
		for i, l := range ls {
			pl, err := lineFrom(l)
			if err != nil {
				return nil, err
			}
			out[i] = pl
		}
		return out, nil
	case "Polygon":
		var rings [][][]float64
		if err := decode(&rings); err != nil {
			return nil, err
		}
		return polygonFrom(rings)
	case "MultiPolygon":
		var polys [][][][]float64
		if err := decode(&polys); err != nil {
			return nil, err
		}
		out := make(MultiPolygon, len(polys))
		for i, rings := range polys {
			p, err := polygonFrom(rings)
			if err != nil {
				return nil, err
			}
			out[i] = p
		}
		return out, nil
	}
	return nil, fmt.Errorf("%w: unsupported geojson type %q", ErrParse, raw.Type)
}

func pointFrom(p []float64) (Coordinate, error) {
	if len(p) < 2 {
		return Coordinate{}, fmt.Errorf("%w: position needs [lng, lat]", ErrParse)
	}
	return New(p[1], p[0])
}

func coordsFrom(ps [][]float64) ([]Coordinate, error) {
	out := make([]Coordinate, len(ps))
	for i, p := range ps {
		c, err := pointFrom(p)
		if err != nil {
			return nil, err
		}
		out[i] = c
	}
	return out, nil
}

func lineFrom(ps [][]float64) (Polyline, error) {
	if len(ps) < 2 {
		return nil, fmt.Errorf("%w: LineString needs 2+ positions", ErrParse)
	}
	cs, err := coordsFrom(ps)
	return Polyline(cs), err
}

func polygonFrom(rings [][][]float64) (PolygonWithHoles, error) {
	if len(rings) == 0 {
		return PolygonWithHoles{}, fmt.Errorf("%w: Polygon needs a ring", ErrParse)
	}
	var out PolygonWithHoles
	for i, r := range rings {
		cs, err := coordsFrom(r)
		if err != nil {
			return PolygonWithHoles{}, err
		}
		p := Polygon(Polygon(cs).ring())
		if err := p.Validate(); err != nil {
			return PolygonWithHoles{}, err
		}
		if i == 0 {
			out.Outer = p
		} else {
			out.Holes = append(out.Holes, p)
		}
	}
	return out, nil
}

// MarshalGeoJSON encodes c as a GeoJSON Point.
func (c Coordinate) MarshalGeoJSON() ([]byte, error) { return MarshalGeometry(c) }

// MarshalGeoJSON encodes the line as a GeoJSON LineString.
func (l Polyline) MarshalGeoJSON() ([]byte, error) { return MarshalGeometry(l) }

// MarshalGeoJSON encodes the polygon as a GeoJSON Polygon with a closed,
// counter-clockwise exterior ring (RFC 7946 right-hand rule).
func (p Polygon) MarshalGeoJSON() ([]byte, error) { return MarshalGeometry(p) }

// ParseGeoJSONPoint decodes a GeoJSON Point geometry.
func ParseGeoJSONPoint(data []byte) (Coordinate, error) {
	g, err := ParseGeometry(data)
	if err != nil {
		return Coordinate{}, err
	}
	c, ok := g.(Coordinate)
	if !ok {
		return Coordinate{}, fmt.Errorf("%w: geojson type %q, want Point", ErrParse, g.GeoJSONType())
	}
	return c, nil
}

// Feature is a GeoJSON Feature. Geometry may be nil.
type Feature struct {
	ID         any
	Geometry   Geometry
	Properties map[string]any
}

type rawFeature struct {
	Type       string          `json:"type"`
	ID         any             `json:"id,omitempty"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties map[string]any  `json:"properties"`
}

func (f Feature) MarshalJSON() ([]byte, error) {
	raw := rawFeature{Type: "Feature", ID: f.ID, Properties: f.Properties, Geometry: json.RawMessage("null")}
	if f.Geometry != nil {
		g, err := MarshalGeometry(f.Geometry)
		if err != nil {
			return nil, err
		}
		raw.Geometry = g
	}
	return json.Marshal(raw)
}

func (f *Feature) UnmarshalJSON(data []byte) error {
	var raw rawFeature
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: %v", ErrParse, err)
	}
	if raw.Type != "Feature" {
		return fmt.Errorf("%w: type %q, want Feature", ErrParse, raw.Type)
	}
	out := Feature{ID: raw.ID, Properties: raw.Properties}
	if len(raw.Geometry) > 0 && string(raw.Geometry) != "null" {
		g, err := ParseGeometry(raw.Geometry)
		if err != nil {
			return err
		}
		out.Geometry = g
	}
	*f = out
	return nil
}

// FeatureCollection is a GeoJSON FeatureCollection.
type FeatureCollection struct {
	Features []Feature
}

type rawFeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

func (fc FeatureCollection) MarshalJSON() ([]byte, error) {
	fs := fc.Features
	if fs == nil {
		fs = []Feature{}
	}
	return json.Marshal(rawFeatureCollection{Type: "FeatureCollection", Features: fs})
}

func (fc *FeatureCollection) UnmarshalJSON(data []byte) error {
	var raw rawFeatureCollection
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Type != "FeatureCollection" {
		return fmt.Errorf("%w: type %q, want FeatureCollection", ErrParse, raw.Type)
	}
	fc.Features = raw.Features
	return nil
}
