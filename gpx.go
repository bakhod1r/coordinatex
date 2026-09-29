package coordinatex

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type gpxPoint struct {
	Lat  float64    `xml:"lat,attr"`
	Lon  float64    `xml:"lon,attr"`
	Time *time.Time `xml:"time,omitempty"`
}

type gpxDoc struct {
	XMLName xml.Name   `xml:"gpx"`
	Version string     `xml:"version,attr,omitempty"`
	Creator string     `xml:"creator,attr,omitempty"`
	Xmlns   string     `xml:"xmlns,attr,omitempty"`
	Wpt     []gpxPoint `xml:"wpt"`
	Rte     []struct {
		Rtept []gpxPoint `xml:"rtept"`
	} `xml:"rte"`
	Trk []gpxTrack `xml:"trk"`
}

type gpxTrack struct {
	Name   string `xml:"name,omitempty"`
	Trkseg []struct {
		Trkpt []gpxPoint `xml:"trkpt"`
	} `xml:"trkseg"`
}

// WriteGPX writes the track as a GPX 1.1 document with one track segment.
func WriteGPX(w io.Writer, name string, t Track) error {
	trk := gpxTrack{Name: name}
	trk.Trkseg = make([]struct {
		Trkpt []gpxPoint `xml:"trkpt"`
	}, 1)
	for _, p := range t {
		if err := p.Validate(); err != nil {
			return err
		}
		gp := gpxPoint{Lat: p.Lat, Lon: p.Lng}
		if !p.Time.IsZero() {
			tm := p.Time.UTC()
			gp.Time = &tm
		}
		trk.Trkseg[0].Trkpt = append(trk.Trkseg[0].Trkpt, gp)
	}
	doc := gpxDoc{Version: "1.1", Creator: "coordinatex", Xmlns: "http://www.topografix.com/GPX/1/1", Trk: []gpxTrack{trk}}
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(doc)
}

// ReadGPX reads waypoints, route points and track points (in that order) from a GPX document.
func ReadGPX(r io.Reader) (Track, error) {
	var doc gpxDoc
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("%w: gpx: %v", ErrParse, err)
	}
	var out Track
	add := func(p gpxPoint) error {
		c, err := New(p.Lat, p.Lon)
		if err != nil {
			return err
		}
		tp := TrackPoint{Coordinate: c}
		if p.Time != nil {
			tp.Time = *p.Time
		}
		out = append(out, tp)
		return nil
	}
	for _, p := range doc.Wpt {
		if err := add(p); err != nil {
			return nil, err
		}
	}
	for _, rte := range doc.Rte {
		for _, p := range rte.Rtept {
			if err := add(p); err != nil {
				return nil, err
			}
		}
	}
	for _, trk := range doc.Trk {
		for _, seg := range trk.Trkseg {
			for _, p := range seg.Trkpt {
				if err := add(p); err != nil {
					return nil, err
				}
			}
		}
	}
	return out, nil
}

// WriteKML writes features as a KML 2.2 document. Supported geometries: Point
// (Coordinate), LineString (Polyline), Polygon and PolygonWithHoles. A string
// "name" property becomes the placemark name.
func WriteKML(w io.Writer, name string, features ...Feature) error {
	var sb strings.Builder
	sb.WriteString(xml.Header)
	sb.WriteString(`<kml xmlns="http://www.opengis.net/kml/2.2"><Document><name>`)
	xmlEscape(&sb, name)
	sb.WriteString("</name>")
	for _, f := range features {
		sb.WriteString("<Placemark>")
		if n, ok := f.Properties["name"].(string); ok {
			sb.WriteString("<name>")
			xmlEscape(&sb, n)
			sb.WriteString("</name>")
		}
		switch g := f.Geometry.(type) {
		case Coordinate:
			sb.WriteString("<Point><coordinates>" + kmlCoords([]Coordinate{g}) + "</coordinates></Point>")
		case Polyline:
			sb.WriteString("<LineString><coordinates>" + kmlCoords(g) + "</coordinates></LineString>")
		case Polygon:
			kmlPolygon(&sb, PolygonWithHoles{Outer: g})
		case PolygonWithHoles:
			kmlPolygon(&sb, g)
		default:
			return fmt.Errorf("%w: KML does not support %T", ErrParse, f.Geometry)
		}
		sb.WriteString("</Placemark>")
	}
	sb.WriteString("</Document></kml>\n")
	_, err := io.WriteString(w, sb.String())
	return err
}

func xmlEscape(sb *strings.Builder, s string) { _ = xml.EscapeText(sb, []byte(s)) }

func kmlCoords(cs []Coordinate) string {
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = strconv.FormatFloat(c.Lng, 'f', -1, 64) + "," + strconv.FormatFloat(c.Lat, 'f', -1, 64)
	}
	return strings.Join(parts, " ")
}

func kmlPolygon(sb *strings.Builder, p PolygonWithHoles) {
	ring := func(r Polygon) string {
		cs := r.ring()
		return "<LinearRing><coordinates>" + kmlCoords(append(append([]Coordinate{}, cs...), cs[0])) + "</coordinates></LinearRing>"
	}
	sb.WriteString("<Polygon><outerBoundaryIs>" + ring(p.Outer) + "</outerBoundaryIs>")
	for _, h := range p.Holes {
		sb.WriteString("<innerBoundaryIs>" + ring(h) + "</innerBoundaryIs>")
	}
	sb.WriteString("</Polygon>")
}
