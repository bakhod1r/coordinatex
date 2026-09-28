package coordinatex

import "testing"

func TestDistanceBetween(t *testing.T) {
	near(t, "Tashkent-Samarkand km", DistanceBetween(tashkent, samarkand).Kilometers(), 266, 2)
	near(t, "London-Paris km", DistanceBetween(london, paris).Kilometers(), 343.5, 1)
	near(t, "same", DistanceBetween(tashkent, tashkent).Meters(), 0, 1e-9)
	near(t, "symmetric", DistanceBetween(paris, london).Meters(), DistanceBetween(london, paris).Meters(), 1e-6)
	// half circumference
	near(t, "antipode", DistanceBetween(Coordinate{0, 0}, Coordinate{0, 180}).Meters(), 3.141592653589793*EarthRadius.Meters(), 1)
}

func TestEquirectangularDistance(t *testing.T) {
	h := DistanceBetween(tashkent, samarkand).Meters()
	e := EquirectangularDistance(tashkent, samarkand).Meters()
	near(t, "equirect ≈ haversine", e, h, h*0.005)
	// across antimeridian must not go the long way
	near(t, "antimeridian", EquirectangularDistance(Coordinate{0, 179.5}, Coordinate{0, -179.5}).Kilometers(), 111.2, 1)
}

func TestDistanceVincenty(t *testing.T) {
	// Flinders Peak - Buninyong (Vincenty 1975 test)
	a := Coordinate{-(37 + 57/60.0 + 3.72030/3600), 144 + 25/60.0 + 29.52440/3600}
	b := Coordinate{-(37 + 39/60.0 + 10.15610/3600), 143 + 55/60.0 + 35.38390/3600}
	d, err := DistanceVincenty(a, b)
	if err != nil {
		t.Fatal(err)
	}
	near(t, "Vincenty", d.Meters(), 54972.271, 0.01)

	d, err = DistanceVincenty(a, a)
	if err != nil || d != 0 {
		t.Fatalf("same point: %v %v", d, err)
	}
	// nearly antipodal: may fail to converge
	if _, err := DistanceVincenty(Coordinate{0, 0}, Coordinate{0.5, 179.7}); err == nil {
		t.Log("converged for near-antipodal (acceptable)")
	}
}
