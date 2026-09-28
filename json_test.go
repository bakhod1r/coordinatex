package coordinatex

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestJSON(t *testing.T) {
	b, err := json.Marshal(Coordinate{41.3111, 69.2797})
	if err != nil || string(b) != `{"lat":41.3111,"lng":69.2797}` {
		t.Fatal(string(b), err)
	}
	var c Coordinate
	if err := json.Unmarshal([]byte(`{"lat":41.3111,"lng":69.2797}`), &c); err != nil || c != (Coordinate{41.3111, 69.2797}) {
		t.Fatal(c, err)
	}
	if err := json.Unmarshal([]byte(`{"lat":91,"lng":0}`), &c); !errors.Is(err, ErrInvalidLatitude) {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"lat":"x"}`), &c); err == nil {
		t.Fatal("bad type accepted")
	}
	if _, err := json.Marshal(Coordinate{100, 0}); err == nil {
		t.Fatal("invalid marshalled")
	}
}
