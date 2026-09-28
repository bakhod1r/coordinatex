package coordinatex

// Fence is any region that can answer point membership.
type Fence interface {
	Contains(Coordinate) bool
}

// FenceFunc adapts a function to Fence.
type FenceFunc func(Coordinate) bool

func (f FenceFunc) Contains(c Coordinate) bool { return f(c) }

// CircleFence contains points within Radius of Center.
type CircleFence struct {
	Center Coordinate
	Radius Distance
}

func (f CircleFence) Contains(c Coordinate) bool { return WithinRadius(f.Center, c, f.Radius) }

// PolygonFence contains points inside Polygon.
type PolygonFence struct{ Polygon Polygon }

func (f PolygonFence) Contains(c Coordinate) bool { return f.Polygon.Contains(c) }

// RectangleFence contains points inside Bounds.
type RectangleFence struct{ Bounds Bounds }

func (f RectangleFence) Contains(c Coordinate) bool { return f.Bounds.Contains(c) }

// AnyFence contains a point if any fence does (union). Empty means nothing.
func AnyFence(fences ...Fence) Fence {
	return FenceFunc(func(c Coordinate) bool {
		for _, f := range fences {
			if f.Contains(c) {
				return true
			}
		}
		return false
	})
}

// AllFence contains a point if every fence does (intersection). Empty means everything.
func AllFence(fences ...Fence) Fence {
	return FenceFunc(func(c Coordinate) bool {
		for _, f := range fences {
			if !f.Contains(c) {
				return false
			}
		}
		return true
	})
}

// Transition describes how a moving point relates to a fence between two fixes.
type Transition int

const (
	StayedOutside Transition = iota
	Entered
	Exited
	StayedInside
)

func (t Transition) String() string {
	switch t {
	case StayedOutside:
		return "stayed_outside"
	case Entered:
		return "entered"
	case Exited:
		return "exited"
	case StayedInside:
		return "stayed_inside"
	}
	return "unknown"
}

// DetectTransition compares the previous and current position against f.
func DetectTransition(f Fence, prev, cur Coordinate) Transition {
	was, is := f.Contains(prev), f.Contains(cur)
	switch {
	case !was && is:
		return Entered
	case was && !is:
		return Exited
	case is:
		return StayedInside
	}
	return StayedOutside
}
