package coordinatex

import (
	"fmt"
	"math"
)

// DistanceBetween returns the great-circle distance using the haversine formula.
func DistanceBetween(a, b Coordinate) Distance {
	φ1, φ2 := toRad(a.Lat), toRad(b.Lat)
	Δφ := φ2 - φ1
	Δλ := toRad(b.Lng - a.Lng)
	h := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	return Distance(2*math.Atan2(math.Sqrt(h), math.Sqrt(1-h))) * EarthRadius
}

// EquirectangularDistance is a fast approximation, accurate for short distances.
func EquirectangularDistance(a, b Coordinate) Distance {
	x := toRad(wrapLng(b.Lng-a.Lng)) * math.Cos(toRad((a.Lat+b.Lat)/2))
	y := toRad(b.Lat - a.Lat)
	return Distance(math.Hypot(x, y)) * EarthRadius
}

// WGS84 ellipsoid parameters.
const (
	wgs84A = 6378137.0
	wgs84F = 1 / 298.257223563
	wgs84B = wgs84A * (1 - wgs84F)
)

// DistanceVincenty returns the geodesic distance on the WGS84 ellipsoid
// (sub-millimetre accuracy). It returns ErrNoConvergence for nearly antipodal points.
func DistanceVincenty(a, b Coordinate) (Distance, error) {
	L := toRad(b.Lng - a.Lng)
	U1 := math.Atan((1 - wgs84F) * math.Tan(toRad(a.Lat)))
	U2 := math.Atan((1 - wgs84F) * math.Tan(toRad(b.Lat)))
	sinU1, cosU1 := math.Sincos(U1)
	sinU2, cosU2 := math.Sincos(U2)

	λ := L
	var sinσ, cosσ, σ, cos2α, cos2σm float64
	for i := 0; i < 200; i++ {
		sinλ, cosλ := math.Sincos(λ)
		sinσ = math.Hypot(cosU2*sinλ, cosU1*sinU2-sinU1*cosU2*cosλ)
		if sinσ == 0 {
			return 0, nil // coincident points
		}
		cosσ = sinU1*sinU2 + cosU1*cosU2*cosλ
		σ = math.Atan2(sinσ, cosσ)
		sinα := cosU1 * cosU2 * sinλ / sinσ
		cos2α = 1 - sinα*sinα
		if cos2α != 0 {
			cos2σm = cosσ - 2*sinU1*sinU2/cos2α
		} else {
			cos2σm = 0 // equatorial line
		}
		C := wgs84F / 16 * cos2α * (4 + wgs84F*(4-3*cos2α))
		prev := λ
		λ = L + (1-C)*wgs84F*sinα*(σ+C*sinσ*(cos2σm+C*cosσ*(-1+2*cos2σm*cos2σm)))
		if math.Abs(λ-prev) < 1e-12 {
			u2 := cos2α * (wgs84A*wgs84A - wgs84B*wgs84B) / (wgs84B * wgs84B)
			A := 1 + u2/16384*(4096+u2*(-768+u2*(320-175*u2)))
			B := u2 / 1024 * (256 + u2*(-128+u2*(74-47*u2)))
			Δσ := B * sinσ * (cos2σm + B/4*(cosσ*(-1+2*cos2σm*cos2σm)-B/6*cos2σm*(-3+4*sinσ*sinσ)*(-3+4*cos2σm*cos2σm)))
			return Distance(wgs84B * A * (σ - Δσ)), nil
		}
	}
	return 0, fmt.Errorf("%w: vincenty %v -> %v", ErrNoConvergence, a, b)
}

// DestinationVincenty solves the direct geodesic problem on the WGS84 ellipsoid:
// the point reached from start after distance d along the initial bearing.
func DestinationVincenty(start Coordinate, d Distance, bearing Angle) (Coordinate, error) {
	if d == 0 {
		return start, nil
	}
	α1 := bearing.Radians()
	sinα1, cosα1 := math.Sincos(α1)
	tanU1 := (1 - wgs84F) * math.Tan(toRad(start.Lat))
	cosU1 := 1 / math.Sqrt(1+tanU1*tanU1)
	sinU1 := tanU1 * cosU1
	σ1 := math.Atan2(tanU1, cosα1)
	sinα := cosU1 * sinα1
	cos2α := 1 - sinα*sinα
	u2 := cos2α * (wgs84A*wgs84A - wgs84B*wgs84B) / (wgs84B * wgs84B)
	A := 1 + u2/16384*(4096+u2*(-768+u2*(320-175*u2)))
	B := u2 / 1024 * (256 + u2*(-128+u2*(74-47*u2)))

	σ := float64(d) / (wgs84B * A)
	var sinσ, cosσ, cos2σm float64
	converged := false
	for i := 0; i < 200; i++ {
		cos2σm = math.Cos(2*σ1 + σ)
		sinσ, cosσ = math.Sincos(σ)
		Δσ := B * sinσ * (cos2σm + B/4*(cosσ*(-1+2*cos2σm*cos2σm)-B/6*cos2σm*(-3+4*sinσ*sinσ)*(-3+4*cos2σm*cos2σm)))
		prev := σ
		σ = float64(d)/(wgs84B*A) + Δσ
		if math.Abs(σ-prev) < 1e-12 {
			converged = true
			break
		}
	}
	if !converged {
		return Coordinate{}, fmt.Errorf("%w: vincenty direct", ErrNoConvergence)
	}
	cos2σm = math.Cos(2*σ1 + σ)
	sinσ, cosσ = math.Sincos(σ)
	x := sinU1*sinσ - cosU1*cosσ*cosα1
	φ2 := math.Atan2(sinU1*cosσ+cosU1*sinσ*cosα1, (1-wgs84F)*math.Hypot(sinα, x))
	λ := math.Atan2(sinσ*sinα1, cosU1*cosσ-sinU1*sinσ*cosα1)
	C := wgs84F / 16 * cos2α * (4 + wgs84F*(4-3*cos2α))
	L := λ - (1-C)*wgs84F*sinα*(σ+C*sinσ*(cos2σm+C*cosσ*(-1+2*cos2σm*cos2σm)))
	return Coordinate{Lat: toDeg(φ2), Lng: wrapLng(start.Lng + toDeg(L))}, nil
}
