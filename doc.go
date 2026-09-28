// Package coordinatex provides small, strongly-typed geospatial primitives
// for Go backend services: coordinates, distances, bearings, bounding boxes,
// geohashes, polygons, polylines, geofences and common serialization formats.
//
// All calculations use a spherical Earth model (mean radius, IUGG) unless the
// function name says otherwise (for example DistanceVincenty, which uses the
// WGS84 ellipsoid). Coordinates are always (latitude, longitude) in decimal
// degrees; formats that use (longitude, latitude) ordering such as GeoJSON and
// WKT are converted explicitly.
package coordinatex
