package coordinatex

import (
	"math"
	"sort"
)

// GridIndex is an in-memory uniform-grid spatial index for radius and nearest
// queries. It is not safe for concurrent writes; guard it with a sync.RWMutex
// if shared. For large or persistent datasets prefer PostGIS or a geohash column.
type GridIndex[T comparable] struct {
	cellDeg float64
	cells   map[[2]int64]map[T]Coordinate
	pos     map[T]Coordinate
}

// NewGridIndex creates an index with roughly cell-sized grid squares. Pick a
// cell close to your typical query radius.
func NewGridIndex[T comparable](cell Distance) *GridIndex[T] {
	return &GridIndex[T]{
		cellDeg: toDeg(float64(cell / EarthRadius)),
		cells:   map[[2]int64]map[T]Coordinate{},
		pos:     map[T]Coordinate{},
	}
}

func (g *GridIndex[T]) key(lat, lng float64) [2]int64 {
	return [2]int64{int64(math.Floor(lat / g.cellDeg)), int64(math.Floor((lng + 180) / g.cellDeg))}
}

// Insert adds or moves id to c.
func (g *GridIndex[T]) Insert(id T, c Coordinate) {
	g.Remove(id)
	k := g.key(c.Lat, c.Lng)
	if g.cells[k] == nil {
		g.cells[k] = map[T]Coordinate{}
	}
	g.cells[k][id] = c
	g.pos[id] = c
}

// Remove deletes id; unknown ids are ignored.
func (g *GridIndex[T]) Remove(id T) {
	c, ok := g.pos[id]
	if !ok {
		return
	}
	k := g.key(c.Lat, c.Lng)
	delete(g.cells[k], id)
	if len(g.cells[k]) == 0 {
		delete(g.cells, k)
	}
	delete(g.pos, id)
}

// Len returns the number of indexed items.
func (g *GridIndex[T]) Len() int { return len(g.pos) }

// WithinRadius returns ids within r of center, nearest first.
func (g *GridIndex[T]) WithinRadius(center Coordinate, r Distance) []T {
	type hit struct {
		id T
		d  Distance
	}
	var hits []hit
	visit := func(id T, c Coordinate) {
		if d := DistanceBetween(center, c); d <= r {
			hits = append(hits, hit{id, d})
		}
	}
	b := BoundsAround(center, r)
	rowLo, rowHi := int64(math.Floor(b.MinLat/g.cellDeg)), int64(math.Floor(b.MaxLat/g.cellDeg))
	var cols [][2]int64
	n := 0
	for _, iv := range b.lngIntervals() {
		lo, hi := int64(math.Floor((iv[0]+180)/g.cellDeg)), int64(math.Floor((iv[1]+180)/g.cellDeg))
		cols = append(cols, [2]int64{lo, hi})
		n += int(hi - lo + 1)
	}
	if n*int(rowHi-rowLo+1) > len(g.pos) { // scanning cells would cost more than a linear scan
		for id, c := range g.pos {
			visit(id, c)
		}
	} else {
		for row := rowLo; row <= rowHi; row++ {
			for _, cr := range cols {
				for col := cr[0]; col <= cr[1]; col++ {
					for id, c := range g.cells[[2]int64{row, col}] {
						visit(id, c)
					}
				}
			}
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].d < hits[j].d })
	out := make([]T, len(hits))
	for i, h := range hits {
		out[i] = h.id
	}
	return out
}

// Nearest returns the closest id within maxRadius.
func (g *GridIndex[T]) Nearest(center Coordinate, maxRadius Distance) (T, bool) {
	if hits := g.WithinRadius(center, maxRadius); len(hits) > 0 {
		return hits[0], true
	}
	var zero T
	return zero, false
}
