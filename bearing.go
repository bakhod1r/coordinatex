package coordinatex

import "math"

// CompassPoint is one of the eight principal wind directions.
type CompassPoint string

const (
	North     CompassPoint = "N"
	NorthEast CompassPoint = "NE"
	East      CompassPoint = "E"
	SouthEast CompassPoint = "SE"
	South     CompassPoint = "S"
	SouthWest CompassPoint = "SW"
	West      CompassPoint = "W"
	NorthWest CompassPoint = "NW"
)

var compassPoints = [8]CompassPoint{North, NorthEast, East, SouthEast, South, SouthWest, West, NorthWest}

// Bearing returns the initial great-circle bearing from -> to, in [0, 360),
// where 0 is north and 90 is east.
func Bearing(from, to Coordinate) Angle {
	φ1, φ2 := toRad(from.Lat), toRad(to.Lat)
	Δλ := toRad(to.Lng - from.Lng)
	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	return Angle(toDeg(math.Atan2(y, x))).Normalize()
}

// FinalBearing returns the bearing on arrival at to, in [0, 360).
func FinalBearing(from, to Coordinate) Angle {
	return (Bearing(to, from) + 180).Normalize()
}

// Direction returns the 8-point compass direction from -> to.
func Direction(from, to Coordinate) CompassPoint {
	return CompassFromAngle(Bearing(from, to))
}

// CompassFromAngle maps a bearing to the nearest of the 8 compass points.
func CompassFromAngle(a Angle) CompassPoint {
	return compassPoints[int(math.Floor(float64(a.Normalize()+22.5)/45))%8]
}

var compass16 = [16]CompassPoint{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}

// Compass4 maps a bearing to N, E, S or W.
func Compass4(a Angle) CompassPoint {
	return [4]CompassPoint{North, East, South, West}[int(math.Floor(float64(a.Normalize()+45)/90))%4]
}

// Compass16 maps a bearing to one of 16 points (N, NNE, NE, ENE, ...).
func Compass16(a Angle) CompassPoint {
	return compass16[int(math.Floor(float64(a.Normalize()+11.25)/22.5))%16]
}
