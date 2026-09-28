package coordinatex

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// CoordinateFromQuery reads and validates a coordinate from URL query parameters,
// e.g. CoordinateFromQuery(r.URL.Query(), "lat", "lng").
func CoordinateFromQuery(q url.Values, latKey, lngKey string) (Coordinate, error) {
	lat, err := floatParam(q, latKey)
	if err != nil {
		return Coordinate{}, err
	}
	lng, err := floatParam(q, lngKey)
	if err != nil {
		return Coordinate{}, err
	}
	return New(lat, lng)
}

// RadiusFromQuery reads a radius in meters, requiring 0 < radius <= max.
func RadiusFromQuery(q url.Values, key string, max Distance) (Distance, error) {
	v, err := floatParam(q, key)
	if err != nil {
		return 0, err
	}
	if !(v > 0) || Distance(v) > max {
		return 0, fmt.Errorf("%w: %s=%v must be in (0, %v]", ErrParse, key, v, max.Meters())
	}
	return Distance(v), nil
}

// BoundsFromQuery reads "minLng,minLat,maxLng,maxLat" (GeoJSON/OGC bbox order).
func BoundsFromQuery(q url.Values, key string) (Bounds, error) {
	parts := strings.Split(q.Get(key), ",")
	if len(parts) != 4 {
		return Bounds{}, fmt.Errorf("%w: %s needs minLng,minLat,maxLng,maxLat", ErrParse, key)
	}
	var v [4]float64
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return Bounds{}, fmt.Errorf("%w: %s[%d]=%q", ErrParse, key, i, p)
		}
		v[i] = f
	}
	return NewBounds(Coordinate{Lat: v[1], Lng: v[0]}, Coordinate{Lat: v[3], Lng: v[2]})
}

func floatParam(q url.Values, key string) (float64, error) {
	s := strings.TrimSpace(q.Get(key))
	if s == "" {
		return 0, fmt.Errorf("%w: missing %s", ErrParse, key)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s=%q", ErrParse, key, s)
	}
	return f, nil
}
