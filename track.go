package coordinatex

import (
	"math"
	"time"
)

// TrackPoint is a timestamped GPS fix.
type TrackPoint struct {
	Coordinate
	Time     time.Time `json:"time"`
	Accuracy Distance  `json:"accuracy,omitempty"` // horizontal accuracy radius, 0 if unknown
}

// Track is a time-ordered sequence of fixes.
type Track []TrackPoint

// Polyline drops timestamps.
func (t Track) Polyline() Polyline {
	out := make(Polyline, len(t))
	for i, p := range t {
		out[i] = p.Coordinate
	}
	return out
}

// Distance returns the travelled distance.
func (t Track) Distance() Distance { return t.Polyline().Length() }

// Duration returns the time between first and last fix.
func (t Track) Duration() time.Duration {
	if len(t) < 2 {
		return 0
	}
	return t[len(t)-1].Time.Sub(t[0].Time)
}

// AverageSpeed is Distance / Duration, 0 when duration is not positive.
func (t Track) AverageSpeed() Speed {
	d := t.Duration()
	if d <= 0 {
		return 0
	}
	return Speed(t.Distance().Meters() / d.Seconds())
}

// Speeds returns the speed of each segment (0 where time does not advance).
func (t Track) Speeds() []Speed {
	if len(t) < 2 {
		return nil
	}
	out := make([]Speed, len(t)-1)
	for i := range out {
		out[i] = segmentSpeed(t[i], t[i+1])
	}
	return out
}

// MaxSpeed returns the fastest segment speed.
func (t Track) MaxSpeed() Speed {
	var m Speed
	for _, s := range t.Speeds() {
		m = Speed(math.Max(float64(m), float64(s)))
	}
	return m
}

// FilterSpeed removes fixes that would require moving faster than max from the
// last kept fix, and fixes whose time does not advance (GPS spikes and duplicates).
func (t Track) FilterSpeed(max Speed) Track {
	if len(t) == 0 {
		return nil
	}
	out := Track{t[0]}
	for _, p := range t[1:] {
		last := out[len(out)-1]
		if !p.Time.After(last.Time) || segmentSpeed(last, p) > max {
			continue
		}
		out = append(out, p)
	}
	return out
}

// IsStationary reports whether every fix is within radius of the first one.
func (t Track) IsStationary(radius Distance) bool {
	for _, p := range t {
		if !WithinRadius(t[0].Coordinate, p.Coordinate, radius) {
			return false
		}
	}
	return true
}

func segmentSpeed(a, b TrackPoint) Speed {
	dt := b.Time.Sub(a.Time).Seconds()
	if dt <= 0 {
		return 0
	}
	return Speed(DistanceBetween(a.Coordinate, b.Coordinate).Meters() / dt)
}

// Smooth applies a constant-velocity Kalman filter to the track. accelNoise is
// the expected acceleration in m/s² (≈1 for walking/driving at steady speed,
// higher for erratic motion). Each fix's Accuracy is the measurement error
// (10 m if unset); the output Accuracy is the filter's position uncertainty.
func (t Track) Smooth(accelNoise float64) Track {
	if len(t) == 0 {
		return nil
	}
	lat0, lng0 := t[0].Lat, t[0].Lng
	cosφ := math.Cos(toRad(lat0))
	R := float64(EarthRadius)
	toXY := func(c Coordinate) (float64, float64) {
		return toRad(wrapLng(c.Lng-lng0)) * cosφ * R, toRad(c.Lat-lat0) * R
	}
	acc := func(p TrackPoint) float64 {
		if p.Accuracy > 0 {
			return float64(p.Accuracy)
		}
		return 10
	}
	type axis struct{ p, v, p00, p01, p11 float64 }
	x0, y0 := toXY(t[0].Coordinate)
	r0 := acc(t[0]) * acc(t[0])
	ax := [2]axis{{p: x0, p00: r0, p11: 1000}, {p: y0, p00: r0, p11: 1000}}
	q2 := accelNoise * accelNoise

	out := make(Track, len(t))
	out[0] = t[0]
	for i := 1; i < len(t); i++ {
		dt := t[i].Time.Sub(t[i-1].Time).Seconds()
		if dt < 0 {
			dt = 0
		}
		zx, zy := toXY(t[i].Coordinate)
		r := acc(t[i]) * acc(t[i])
		for k, z := range [2]float64{zx, zy} {
			a := &ax[k]
			// predict
			a.p += a.v * dt
			a.p00 += dt*(2*a.p01+dt*a.p11) + q2*dt*dt*dt*dt/4
			a.p01 += dt*a.p11 + q2*dt*dt*dt/2
			a.p11 += q2 * dt * dt
			// update
			s := a.p00 + r
			k0, k1 := a.p00/s, a.p01/s
			y := z - a.p
			a.p += k0 * y
			a.v += k1 * y
			a.p11 -= k1 * a.p01
			a.p01 -= k0 * a.p01
			a.p00 -= k0 * a.p00
		}
		out[i] = t[i]
		out[i].Coordinate = Coordinate{
			Lat: lat0 + toDeg(ax[1].p/R),
			Lng: wrapLng(lng0 + toDeg(ax[0].p/(R*cosφ))),
		}
		out[i].Accuracy = Distance(math.Sqrt((ax[0].p00 + ax[1].p00) / 2))
	}
	return out
}
