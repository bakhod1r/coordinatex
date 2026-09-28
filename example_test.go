package coordinatex_test

import (
	"fmt"

	"github.com/bakhod1r/coordinatex"
)

func Example() {
	tashkent := coordinatex.MustNew(41.2995, 69.2401)
	samarkand := coordinatex.MustNew(39.6542, 66.9597)

	d := coordinatex.DistanceBetween(tashkent, samarkand)
	fmt.Printf("%.0f km %s\n", d.Kilometers(), coordinatex.Direction(tashkent, samarkand))
	// Output: 266 km SW
}

func ExampleWithinRadius() {
	user := coordinatex.MustNew(41.3111, 69.2797)
	shop := coordinatex.Destination(user, 3*coordinatex.Kilometer, 45)
	fmt.Println(coordinatex.WithinRadius(user, shop, 5*coordinatex.Kilometer))
	// Output: true
}

func ExampleBoundsAround() {
	b := coordinatex.BoundsAround(coordinatex.MustNew(41.3111, 69.2797), 5*coordinatex.Kilometer)
	// SELECT ... WHERE lat BETWEEN $1 AND $2 AND lng BETWEEN $3 AND $4
	fmt.Printf("%.4f %.4f %.4f %.4f\n", b.MinLat, b.MaxLat, b.MinLng, b.MaxLng)
	// Output: 41.2661 41.3561 69.2198 69.3396
}

func ExampleParse() {
	c, _ := coordinatex.Parse(`41°18'39.96"N 69°16'46.92"E`)
	fmt.Println(c.FormatDecimal(4), c.FormatDM())
	// Output: 41.3111,69.2797 41°18.666'N 69°16.782'E
}

func ExampleEncodeGeohash() {
	fmt.Println(coordinatex.EncodeGeohash(coordinatex.MustNew(57.64911, 10.40744), 11))
	// Output: u4pruydqqvj
}

func ExampleDetectTransition() {
	fence := coordinatex.CircleFence{Center: coordinatex.MustNew(41.3111, 69.2797), Radius: 500}
	fmt.Println(coordinatex.DetectTransition(fence, coordinatex.MustNew(41.33, 69.30), coordinatex.MustNew(41.3112, 69.2798)))
	// Output: entered
}
