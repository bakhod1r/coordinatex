package coordinatex

import "errors"

var (
	// ErrInvalidLatitude is returned when a latitude is outside [-90, 90].
	ErrInvalidLatitude = errors.New("coordinatex: latitude out of range [-90, 90]")
	// ErrInvalidLongitude is returned when a longitude is outside [-180, 180].
	ErrInvalidLongitude = errors.New("coordinatex: longitude out of range [-180, 180]")
	// ErrNaN is returned when a component is NaN or infinite.
	ErrNaN = errors.New("coordinatex: coordinate component is NaN or Inf")
	// ErrInvalidBounds is returned when south-west is north of north-east.
	ErrInvalidBounds = errors.New("coordinatex: invalid bounds")
	// ErrEmpty is returned when an operation needs at least one point.
	ErrEmpty = errors.New("coordinatex: no points")
	// ErrInvalidGeohash is returned for malformed geohash strings.
	ErrInvalidGeohash = errors.New("coordinatex: invalid geohash")
	// ErrParse is returned when a coordinate string cannot be parsed.
	ErrParse = errors.New("coordinatex: cannot parse coordinate")
	// ErrNoConvergence is returned when an iterative algorithm fails to converge.
	ErrNoConvergence = errors.New("coordinatex: algorithm did not converge")
	// ErrInvalidPolyline is returned for malformed encoded polylines.
	ErrInvalidPolyline = errors.New("coordinatex: invalid encoded polyline")
)

// ErrInvalidPolygon is returned when a polygon has fewer than 3 distinct vertices.
var ErrInvalidPolygon = errors.New("coordinatex: polygon needs at least 3 vertices")
