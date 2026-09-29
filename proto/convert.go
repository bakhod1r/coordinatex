// Package coordinatexproto converts between coordinatex types and the
// coordinatex.v1 protobuf messages (package coordinatexpb). It lives in its own
// module so the core package stays dependency-free.
package coordinatexproto

import (
	"errors"
	"fmt"

	"github.com/bakhod1r/coordinatex"
	pb "github.com/bakhod1r/coordinatex/proto/coordinatexpb"
)

// ErrNilMessage is returned when a required message is nil.
var ErrNilMessage = errors.New("coordinatexproto: nil message")

// FromCoordinate converts to a LatLng message.
func FromCoordinate(c coordinatex.Coordinate) *pb.LatLng { return &pb.LatLng{Lat: c.Lat, Lng: c.Lng} }

// ToCoordinate converts and validates a LatLng message.
func ToCoordinate(m *pb.LatLng) (coordinatex.Coordinate, error) {
	if m == nil {
		return coordinatex.Coordinate{}, ErrNilMessage
	}
	return coordinatex.New(m.GetLat(), m.GetLng())
}

// FromBounds converts to a Bounds message.
func FromBounds(b coordinatex.Bounds) *pb.Bounds {
	return &pb.Bounds{MinLat: b.MinLat, MinLng: b.MinLng, MaxLat: b.MaxLat, MaxLng: b.MaxLng}
}

// ToBounds converts and validates a Bounds message.
func ToBounds(m *pb.Bounds) (coordinatex.Bounds, error) {
	if m == nil {
		return coordinatex.Bounds{}, ErrNilMessage
	}
	return coordinatex.NewBounds(
		coordinatex.Coordinate{Lat: m.GetMinLat(), Lng: m.GetMinLng()},
		coordinatex.Coordinate{Lat: m.GetMaxLat(), Lng: m.GetMaxLng()},
	)
}

// FromCircle converts to a Circle message.
func FromCircle(c coordinatex.Circle) *pb.Circle {
	return &pb.Circle{Center: FromCoordinate(c.Center), RadiusMeters: c.Radius.Meters()}
}

// ToCircle converts and validates a Circle message (radius must be >= 0).
func ToCircle(m *pb.Circle) (coordinatex.Circle, error) {
	if m == nil {
		return coordinatex.Circle{}, ErrNilMessage
	}
	center, err := ToCoordinate(m.GetCenter())
	if err != nil {
		return coordinatex.Circle{}, err
	}
	if !(m.GetRadiusMeters() >= 0) {
		return coordinatex.Circle{}, fmt.Errorf("coordinatexproto: invalid radius %v", m.GetRadiusMeters())
	}
	return coordinatex.Circle{Center: center, Radius: coordinatex.Distance(m.GetRadiusMeters())}, nil
}

func fromPoints(cs []coordinatex.Coordinate) []*pb.LatLng {
	out := make([]*pb.LatLng, len(cs))
	for i, c := range cs {
		out[i] = FromCoordinate(c)
	}
	return out
}

func toPoints(ms []*pb.LatLng) ([]coordinatex.Coordinate, error) {
	out := make([]coordinatex.Coordinate, len(ms))
	for i, m := range ms {
		c, err := ToCoordinate(m)
		if err != nil {
			return nil, fmt.Errorf("point %d: %w", i, err)
		}
		out[i] = c
	}
	return out, nil
}

// FromPolyline converts to a Polyline message.
func FromPolyline(l coordinatex.Polyline) *pb.Polyline { return &pb.Polyline{Points: fromPoints(l)} }

// ToPolyline converts and validates a Polyline message.
func ToPolyline(m *pb.Polyline) (coordinatex.Polyline, error) {
	if m == nil {
		return nil, ErrNilMessage
	}
	cs, err := toPoints(m.GetPoints())
	return coordinatex.Polyline(cs), err
}

// FromPolygon converts to a Polygon message.
func FromPolygon(p coordinatex.PolygonWithHoles) *pb.Polygon {
	out := &pb.Polygon{Outer: &pb.Ring{Points: fromPoints(p.Outer)}}
	for _, h := range p.Holes {
		out.Holes = append(out.Holes, &pb.Ring{Points: fromPoints(h)})
	}
	return out
}

// ToPolygon converts and validates a Polygon message.
func ToPolygon(m *pb.Polygon) (coordinatex.PolygonWithHoles, error) {
	if m == nil {
		return coordinatex.PolygonWithHoles{}, ErrNilMessage
	}
	outer, err := toPoints(m.GetOuter().GetPoints())
	if err != nil {
		return coordinatex.PolygonWithHoles{}, err
	}
	p := coordinatex.PolygonWithHoles{Outer: outer}
	for i, h := range m.GetHoles() {
		cs, err := toPoints(h.GetPoints())
		if err != nil {
			return coordinatex.PolygonWithHoles{}, fmt.Errorf("hole %d: %w", i, err)
		}
		p.Holes = append(p.Holes, cs)
	}
	if err := p.Validate(); err != nil {
		return coordinatex.PolygonWithHoles{}, err
	}
	return p, nil
}
