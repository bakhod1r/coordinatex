package coordinatex

import "encoding/json"

type coordinateJSON struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// MarshalJSON encodes as {"lat":..,"lng":..}, rejecting invalid coordinates.
func (c Coordinate) MarshalJSON() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(coordinateJSON(c))
}

// UnmarshalJSON decodes {"lat":..,"lng":..} and validates the result.
func (c *Coordinate) UnmarshalJSON(data []byte) error {
	var v coordinateJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if err := Coordinate(v).Validate(); err != nil {
		return err
	}
	*c = Coordinate(v)
	return nil
}
