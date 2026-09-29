package coordinatex

import _ "embed"

// OpenAPIComponents is an OpenAPI 3.1 document (JSON) with reusable schemas
// (Coordinate, Bounds, GeoJSON geometries and features) and query parameters
// (lat, lng, radius, bbox) matching this package's encodings. Serve it next to
// your API spec or copy it into your own.
//
//go:embed openapi/components.json
var OpenAPIComponents []byte
