package coordinatex

import (
	"encoding/json"
	"testing"
)

func TestOpenAPIComponents(t *testing.T) {
	var doc struct {
		OpenAPI    string `json:"openapi"`
		Components struct {
			Schemas    map[string]json.RawMessage `json:"schemas"`
			Parameters map[string]struct {
				Name string `json:"name"`
				In   string `json:"in"`
			} `json:"parameters"`
		} `json:"components"`
	}
	if err := json.Unmarshal(OpenAPIComponents, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.OpenAPI != "3.1.0" {
		t.Fatal(doc.OpenAPI)
	}
	for _, s := range []string{"Coordinate", "Bounds", "GeoJSONPosition", "GeoJSONPoint", "GeoJSONLineString", "GeoJSONPolygon", "GeoJSONGeometry", "GeoJSONFeature", "GeoJSONFeatureCollection"} {
		if _, ok := doc.Components.Schemas[s]; !ok {
			t.Errorf("missing schema %s", s)
		}
	}
	for key, name := range map[string]string{"Latitude": "lat", "Longitude": "lng", "Radius": "radius", "BBox": "bbox"} {
		p, ok := doc.Components.Parameters[key]
		if !ok || p.Name != name || p.In != "query" {
			t.Errorf("parameter %s: %+v", key, p)
		}
	}

	// the schema's example must be what the package actually produces
	var coord struct {
		Example json.RawMessage `json:"example"`
	}
	_ = json.Unmarshal(doc.Components.Schemas["Coordinate"], &coord)
	var c Coordinate
	if err := json.Unmarshal(coord.Example, &c); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(c)
	var a, e any
	_ = json.Unmarshal(b, &a)
	_ = json.Unmarshal(coord.Example, &e)
	if string(mustJSON(a)) != string(mustJSON(e)) {
		t.Fatalf("example %s vs marshal %s", coord.Example, b)
	}
	var fc struct {
		Example json.RawMessage `json:"example"`
	}
	_ = json.Unmarshal(doc.Components.Schemas["GeoJSONFeatureCollection"], &fc)
	var parsed FeatureCollection
	if err := json.Unmarshal(fc.Example, &parsed); err != nil || len(parsed.Features) == 0 {
		t.Fatal("feature collection example", err)
	}
}

func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
