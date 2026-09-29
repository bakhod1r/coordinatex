package coordinatex

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// UTM is a Universal Transverse Mercator position on WGS84.
type UTM struct {
	Zone     int     // 1..60
	Band     byte    // latitude band letter C..X
	North    bool    // hemisphere
	Easting  float64 // meters
	Northing float64 // meters
}

const (
	utmK0     = 0.9996
	utmFalseE = 500000.0
	utmFalseN = 10000000.0
	utmBands  = "CDEFGHJKLMNPQRSTUVWXX"
	mgrsBands = "CDEFGHJKLMNPQRSTUVWX"
)

var (
	utmN = wgs84F / (2 - wgs84F)
	utmE = math.Sqrt(wgs84F * (2 - wgs84F))
	utmA = func() float64 {
		n := utmN
		return wgs84A / (1 + n) * (1 + n*n/4 + n*n*n*n/64 + n*n*n*n*n*n/256)
	}()
	utmAlpha = func() [7]float64 {
		n := utmN
		n2, n3, n4, n5, n6 := n*n, n*n*n, n*n*n*n, n*n*n*n*n, n*n*n*n*n*n
		return [7]float64{0,
			n/2 - 2.0/3*n2 + 5.0/16*n3 + 41.0/180*n4 - 127.0/288*n5 + 7891.0/37800*n6,
			13.0/48*n2 - 3.0/5*n3 + 557.0/1440*n4 + 281.0/630*n5 - 1983433.0/1935360*n6,
			61.0/240*n3 - 103.0/140*n4 + 15061.0/26880*n5 + 167603.0/181440*n6,
			49561.0/161280*n4 - 179.0/168*n5 + 6601661.0/7257600*n6,
			34729.0/80640*n5 - 3418889.0/1995840*n6,
			212378941.0 / 319334400 * n6}
	}()
	utmBeta = func() [7]float64 {
		n := utmN
		n2, n3, n4, n5, n6 := n*n, n*n*n, n*n*n*n, n*n*n*n*n, n*n*n*n*n*n
		return [7]float64{0,
			n/2 - 2.0/3*n2 + 37.0/96*n3 - 1.0/360*n4 - 81.0/512*n5 + 96199.0/604800*n6,
			1.0/48*n2 + 1.0/15*n3 - 437.0/1440*n4 + 46.0/105*n5 - 1118711.0/3870720*n6,
			17.0/480*n3 - 37.0/840*n4 - 209.0/4480*n5 + 5569.0/90720*n6,
			4397.0/161280*n4 - 11.0/504*n5 - 830251.0/7257600*n6,
			4583.0/161280*n5 - 108847.0/3991680*n6,
			20648693.0 / 638668800 * n6}
	}()
)

func utmZone(c Coordinate) int {
	lng := wrapLng(c.Lng)
	z := int(math.Floor((lng+180)/6)) + 1
	switch {
	case c.Lat >= 56 && c.Lat < 64 && lng >= 3 && lng < 12:
		z = 32
	case c.Lat >= 72 && c.Lat <= 84 && lng >= 0 && lng < 42:
		z = [4]int{31, 33, 35, 37}[int(math.Floor((lng+3)/12))]
	}
	return z
}

func centralMeridian(zone int) float64 { return float64((zone-1)*6 - 180 + 3) }

// ToUTM projects c (latitude -80..84) with the Karney/Krüger series (mm accuracy),
// including the Norway and Svalbard zone exceptions.
func ToUTM(c Coordinate) (UTM, error) {
	if err := c.Validate(); err != nil {
		return UTM{}, err
	}
	if c.Lat < -80 || c.Lat > 84 {
		return UTM{}, fmt.Errorf("%w: UTM latitude %v not in [-80, 84]", ErrOutOfRange, c.Lat)
	}
	zone := utmZone(c)
	e, n := utmForward(c, zone)
	if c.Lat < 0 {
		n += utmFalseN
	}
	return UTM{Zone: zone, Band: utmBands[int(math.Floor(c.Lat/8+10))], North: c.Lat >= 0, Easting: e, Northing: n}, nil
}

// utmForward returns easting (with false easting) and northing (without false northing).
func utmForward(c Coordinate, zone int) (float64, float64) {
	φ := toRad(c.Lat)
	λ := toRad(wrapLng(c.Lng - centralMeridian(zone)))
	cosλ, sinλ := math.Cos(λ), math.Sin(λ)
	τ := math.Tan(φ)
	σ := math.Sinh(utmE * math.Atanh(utmE*τ/math.Sqrt(1+τ*τ)))
	τp := τ*math.Sqrt(1+σ*σ) - σ*math.Sqrt(1+τ*τ)
	ξp := math.Atan2(τp, cosλ)
	ηp := math.Asinh(sinλ / math.Sqrt(τp*τp+cosλ*cosλ))
	ξ, η := ξp, ηp
	for j := 1; j <= 6; j++ {
		ξ += utmAlpha[j] * math.Sin(2*float64(j)*ξp) * math.Cosh(2*float64(j)*ηp)
		η += utmAlpha[j] * math.Cos(2*float64(j)*ξp) * math.Sinh(2*float64(j)*ηp)
	}
	return utmK0*utmA*η + utmFalseE, utmK0 * utmA * ξ
}

// Coordinate converts back to latitude/longitude.
func (u UTM) Coordinate() (Coordinate, error) {
	if u.Zone < 1 || u.Zone > 60 {
		return Coordinate{}, fmt.Errorf("%w: UTM zone %d", ErrOutOfRange, u.Zone)
	}
	y := u.Northing
	if !u.North {
		y -= utmFalseN
	}
	η := (u.Easting - utmFalseE) / (utmK0 * utmA)
	ξ := y / (utmK0 * utmA)
	ξp, ηp := ξ, η
	for j := 1; j <= 6; j++ {
		ξp -= utmBeta[j] * math.Sin(2*float64(j)*ξ) * math.Cosh(2*float64(j)*η)
		ηp -= utmBeta[j] * math.Cos(2*float64(j)*ξ) * math.Sinh(2*float64(j)*η)
	}
	sinhηp := math.Sinh(ηp)
	sinξp, cosξp := math.Sincos(ξp)
	τp := sinξp / math.Sqrt(sinhηp*sinhηp+cosξp*cosξp)
	τi := τp
	e2 := utmE * utmE
	for i := 0; i < 20; i++ {
		σi := math.Sinh(utmE * math.Atanh(utmE*τi/math.Sqrt(1+τi*τi)))
		τip := τi*math.Sqrt(1+σi*σi) - σi*math.Sqrt(1+τi*τi)
		δ := (τp - τip) / math.Sqrt(1+τip*τip) * (1 + (1-e2)*τi*τi) / ((1 - e2) * math.Sqrt(1+τi*τi))
		τi += δ
		if math.Abs(δ) < 1e-14 {
			break
		}
	}
	lat := toDeg(math.Atan(τi))
	lng := wrapLng(centralMeridian(u.Zone) + toDeg(math.Atan2(sinhηp, cosξp)))
	return New(lat, lng)
}

// String formats as "30U 699316 5710164" (meter precision).
func (u UTM) String() string {
	return fmt.Sprintf("%d%c %.0f %.0f", u.Zone, u.Band, u.Easting, u.Northing)
}

// ParseUTM parses "30U 699316 5710164".
func ParseUTM(s string) (UTM, error) {
	f := strings.Fields(strings.ToUpper(s))
	bad := fmt.Errorf("%w: UTM %q", ErrParse, s)
	if len(f) != 3 || len(f[0]) < 2 {
		return UTM{}, bad
	}
	zone, err := strconv.Atoi(f[0][:len(f[0])-1])
	band := f[0][len(f[0])-1]
	if err != nil || zone < 1 || zone > 60 || strings.IndexByte(mgrsBands, band) < 0 {
		return UTM{}, bad
	}
	e, err1 := strconv.ParseFloat(f[1], 64)
	n, err2 := strconv.ParseFloat(f[2], 64)
	if err1 != nil || err2 != nil {
		return UTM{}, bad
	}
	return UTM{Zone: zone, Band: band, North: band >= 'N', Easting: e, Northing: n}, nil
}

var (
	mgrsCol = [3]string{"ABCDEFGH", "JKLMNPQR", "STUVWXYZ"}
	mgrsRow = [2]string{"ABCDEFGHJKLMNPQRSTUV", "FGHJKLMNPQRSTUVABCDE"}
)

// ToMGRS returns the Military Grid Reference System string for c.
// precision is digits per axis: 5 = 1 m, 4 = 10 m, ... 0 = 100 km square only.
func ToMGRS(c Coordinate, precision int) (string, error) {
	u, err := ToUTM(c)
	if err != nil {
		return "", err
	}
	precision = max(0, min(5, precision))
	col := int(math.Floor(u.Easting / 100000))
	row := int(math.Floor(u.Northing/100000)) % 20
	e := int(math.Floor(math.Mod(u.Easting, 100000)))
	n := int(math.Floor(math.Mod(u.Northing, 100000)))
	div := int(math.Pow10(5 - precision))
	s := fmt.Sprintf("%d%c%c%c", u.Zone, u.Band, mgrsCol[(u.Zone-1)%3][col-1], mgrsRow[(u.Zone-1)%2][row])
	if precision > 0 {
		s += fmt.Sprintf("%0*d%0*d", precision, e/div, precision, n/div)
	}
	return s, nil
}

// ParseMGRS returns the center of the MGRS grid square.
func ParseMGRS(s string) (Coordinate, error) {
	s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	bad := fmt.Errorf("%w: MGRS %q", ErrParse, s)
	i := 0
	for i < len(s) && i < 2 && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || len(s) < i+3 {
		return Coordinate{}, bad
	}
	zone, _ := strconv.Atoi(s[:i])
	band, eL, nL, digits := s[i], s[i+1], s[i+2], s[i+3:]
	bandIdx := strings.IndexByte(mgrsBands, band)
	col := strings.IndexByte(mgrsCol[(max(zone, 1)-1)%3], eL)
	row := strings.IndexByte(mgrsRow[(max(zone, 1)-1)%2], nL)
	if zone < 1 || zone > 60 || bandIdx < 0 || col < 0 || row < 0 || len(digits)%2 != 0 || len(digits) > 10 {
		return Coordinate{}, bad
	}
	p := len(digits) / 2
	var de, dn int
	if p > 0 {
		var err1, err2 error
		de, err1 = strconv.Atoi(digits[:p])
		dn, err2 = strconv.Atoi(digits[p:])
		if err1 != nil || err2 != nil || de < 0 || dn < 0 {
			return Coordinate{}, bad
		}
	}
	cell := math.Pow10(5 - p)
	easting := float64(col+1)*100000 + float64(de)*cell + cell/2
	northing := float64(row)*100000 + float64(dn)*cell + cell/2

	latBand := float64((bandIdx - 10) * 8)
	_, nBand := utmForward(Coordinate{Lat: latBand, Lng: centralMeridian(zone)}, zone)
	if latBand < 0 {
		nBand += utmFalseN
	}
	nBand = math.Floor(nBand/100000) * 100000
	for northing < nBand {
		northing += 2000000
	}
	return UTM{Zone: zone, Band: band, North: band >= 'N', Easting: easting, Northing: northing}.Coordinate()
}
