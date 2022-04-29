package poly

// implements the algorithm described at:
// http://philliplemons.com/posts/ray-casting-algorithm

import (
	"math"
)

type Point struct {
	X, Y float64
}

type Polygon struct {
	points []Point
	edges  [][]Point
	bbox   []float64
}

// New takes as its argument a list of points in clockwise order
func New(points ...Point) *Polygon {
	return &Polygon{
		points: points,
		edges:  make([][]Point, 0),
		bbox:   make([]float64, 0),
	}
}

func (p *Polygon) BoundingBox() (x1, y1, x2, y2 float64) {
	if len(p.bbox) == 4 {
		return p.bbox[0], p.bbox[1], p.bbox[2], p.bbox[3]
	}
	// find min and max x and y
	x1, y1, x2, y2 = p.points[0].X, p.points[0].Y, p.points[0].X, p.points[0].Y
	for _, pt := range p.points {
		if pt.X < x1 {
			x1 = pt.X
		}
		if pt.Y < y1 {
			y1 = pt.Y
		}
		if pt.X >= x2 {
			x2 = pt.X
		}
		if pt.Y >= y2 {
			y2 = pt.Y
		}
	}
	// log.Printf("x1: %.2f, y1: %.2f, x2: %.2f, y2: %.2f", x1, y1, x2, y2)
	p.bbox = []float64{x1, y1, x2, y2}
	return x1, y1, x2, y2
}

func (p *Polygon) BoundingBoxOverlaps(sMinX, sMinY, sMaxX, sMaxY float64) bool {
	rMinX, rMinY, rMaxX, rMaxY := p.BoundingBox()
	return rMinX < sMaxX && sMinX < rMaxX && rMinY < sMaxY && sMinY < rMaxY
}

func (p *Polygon) Edges() [][]Point {
	// avoid doing this work if it's already been done
	if len(p.edges) != 0 {
		return p.edges
	}
	if len(p.points) < 2 {
		panic("bad polygon")
	}
	// iterate over list of points,
	// assigning each pair to an edge,
	// ending with the starting point.
	for i := range p.points {
		p.edges = append(p.edges, []Point{p.points[i], p.points[(i+1)%len(p.points)]})
	}
	return p.edges
}

// Contains tells us whether a test point q is within an arbitrary polygon p
func (p *Polygon) Contains(q Point) bool {
	// note: we can speed this up by doing a rough bounding box check and discarding points obviously outside a rectangular bounding box.
	return p.BoundingBoxOverlaps(q.X, q.Y, q.X, q.Y)

	// begin by assuming q is outside
	isInside := false

	for _, e := range p.Edges() {
		a, b := e[0], e[1]
		// make sure point a is lower than point b
		if a.Y > b.Y {
			a, b = b, a
		}
		// make sure test point is not exact height of edge
		if q.Y == a.Y || q.Y == b.Y {
			q.Y += 0.000001
		}

		// check whether a horizontal ray would intersect
		if q.Y > b.Y || q.Y < a.Y || q.X > math.Max(a.X, b.X) {
			continue // does not intersect
		}

		if q.X < math.Min(a.X, b.X) {
			// note: this algorithm depends on detecting
			// whether a ray passing through the polygon
			// has intersected an even or an odd number of
			// sides. so we just flip our latest result
			// every time we cast a ray through a side.
			//
			// in this case, does intersect, so invert current value
			isInside = !isInside
			continue
		}

		edgeN := b.Y - a.Y
		edgeD := b.X - a.X
		edgeSlope := edgeN / edgeD
		if edgeD == 0 && edgeN != 0 {
			edgeSlope = math.Inf(1)
		}
		pointN := q.Y - a.Y
		pointD := q.X - a.X
		pointSlope := pointN / pointD
		if pointD == 0 && pointN != 0 {
			pointSlope = math.Inf(1)
		}

		if pointSlope >= edgeSlope {
			// does intersect, so invert current value
			isInside = !isInside
			continue
		}
	}

	// done checking edges for intersections
	// so we have detected all rays passing through
	// and the current value of isInside must be
	// false if an even number of edges were intersected
	// and true if odd.
	// we're inside if we've intersected an odd number of edges.
	return isInside
}
