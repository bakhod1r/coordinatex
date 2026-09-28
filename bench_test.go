package coordinatex

import "testing"

func BenchmarkDistanceBetween(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		DistanceBetween(tashkent, samarkand)
	}
}

func BenchmarkEquirectangularDistance(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EquirectangularDistance(tashkent, samarkand)
	}
}

func BenchmarkDistanceVincenty(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = DistanceVincenty(tashkent, samarkand)
	}
}

func BenchmarkPolygonContains(b *testing.B) {
	b.ReportAllocs()
	p := Coordinate{0.5, 0.5}
	for i := 0; i < b.N; i++ {
		square.Contains(p)
	}
}
