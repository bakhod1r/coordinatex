# coordinatex

Small, strongly-typed geospatial primitives for Go backend services. Zero dependencies (stdlib only).

```sh
go get github.com/bakhod1r/coordinatex
```

```go
tashkent := coordinatex.MustNew(41.2995, 69.2401)
samarkand := coordinatex.MustNew(39.6542, 66.9597)

d := coordinatex.DistanceBetween(tashkent, samarkand) // coordinatex.Distance (meters)
fmt.Println(d.Kilometers(), coordinatex.Direction(tashkent, samarkand)) // 266.x SW

if coordinatex.WithinRadius(user, shop, 5*coordinatex.Kilometer) { /* ... */ }
```

## Features

| Area | API |
|---|---|
| Coordinate | `New`, `MustNew`, `Validate`, `IsValid`, `Equal`, `Normalize`, `Round`, `IsZero`, `Antipode`, `String` |
| Units | `Distance` (m, km, cm, mi, ft, yd, nmi), `Area` (m², km², ha, acre), `Angle`, `Speed` (m/s, km/h, mph, kn) |
| Distance | `DistanceBetween` (haversine), `EquirectangularDistance` (fast), `DistanceVincenty` / `DestinationVincenty` (WGS84), `DistanceMatrix` |
| Bearing | `Bearing`, `FinalBearing`, `Angle.Reverse`, `AngleDifference`, `Direction`, `Compass4/8/16` |
| Movement | `Destination`, `Midpoint`, `Interpolate`, `IntermediatePoints`, `PointsEvery` |
| Great circle | `CrossTrackDistance`, `AlongTrackDistance`, `SegmentIntersection` |
| Bounds | `NewBounds`, `BoundsOf`, `BoundsAround`, `Contains`, `Intersects`, `Intersection`, `Union`, `Expand`, `Split`, `Center`, `Width`, `Height`, `Area`, `Perimeter` — antimeridian aware |
| Proximity | `WithinRadius`, `WithinRadiusFilter`, `Nearest`, `NearestN`, `SortByDistance` |
| Geohash | `EncodeGeohash`, `DecodeGeohash`, `GeohashBounds`, `GeohashNeighbors`, `GeohashParent`, `GeohashChildren`, `GeohashPrecisionFor`, `GeohashesAround`, `IsValidGeohash` |
| Polygon | `Validate`, `Contains`, `Area`, `Perimeter`, `Centroid`, `Bounds`, `IsClockwise`, `IsSimple`, `Intersects`, `ContainsPolygon`, `Overlaps`; `PolygonWithHoles` |
| Polyline | `Length`, `Bounds`, `NearestPoint`, `PointAt`, `Deviates`, `Simplify` (Douglas–Peucker), `Reverse`, `SegmentLengths`, `Bearings`, `EncodePolyline`/`DecodePolyline` (Google) |
| Circle | `Contains`, `DistanceTo`, `Intersects`, `Bounds`, `Area`, `Circumference` |
| Geofence | `Fence`, `FenceFunc`, `CircleFence`, `PolygonFence`, `RectangleFence`, `AnyFence`, `AllFence`, `DetectTransition` (entered/exited) |
| GPS | `TrackPoint`, `Track`: `Distance`, `Duration`, `AverageSpeed`, `MaxSpeed`, `Speeds`, `FilterSpeed`, `IsStationary`, `Smooth` (Kalman) |
| Formats | `Parse` (decimal, DMS, DM, hemisphere prefix/suffix), `FormatDMS`, `FormatDM`, `FormatDecimal` |
| Serialization | JSON `{"lat","lng"}` (validated), `MarshalText`/`UnmarshalText`, GeoJSON (all geometry types, `Feature`, `FeatureCollection`, `MarshalGeometry`/`ParseGeometry`), WKT/EWKT, WKB/EWKB (hex + binary), `sql.Scanner`/`driver.Valuer` |
| Index | `GridIndex[T]`: `Insert`, `Remove`, `WithinRadius`, `Nearest` |
| HTTP | `CoordinateFromQuery`, `RadiusFromQuery`, `BoundsFromQuery` |

## Two-stage radius search

```go
b := coordinatex.BoundsAround(center, r)          // 1. index-friendly pre-filter
// SELECT ... WHERE lat BETWEEN b.MinLat AND b.MaxLat AND lng BETWEEN b.MinLng AND b.MaxLng
//   (if b.CrossesAntimeridian(): lng >= b.MinLng OR lng <= b.MaxLng)
hits := coordinatex.WithinRadiusFilter(center, rows, r) // 2. exact check
```

Or with geohash prefixes: `coordinatex.GeohashesAround(center, r)`.

## Conventions

- Coordinates are always `(lat, lng)`; GeoJSON and WKT use `(lng, lat)` and are converted explicitly.
- Spherical Earth (IUGG mean radius 6 371 008.8 m) unless the function says WGS84.
- Polygon `Contains`/`Centroid`/`IsSimple` are planar in lat/lng space — fine for city-scale fences, not for polygons spanning the antimeridian or poles.
- All types are immutable values; every function is safe for concurrent use.

## Roadmap

- GeoJSON bbox/foreign members, 3D positions (Z/M)
- Spherical (not planar) point-in-polygon and polygon relations
- R-tree index for non-point geometries
