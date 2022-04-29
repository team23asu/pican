package demo

import (
	"fmt"
	"math"
)

type Vector2D struct {
	X, Y float64
}

const EqThreshold float64 = 1e-9

// Add returns the Vector2D p + q
func (p Vector2D) Add(q Vector2D) Vector2D {
	return Vector2D{
		X: p.X + q.X,
		Y: p.Y + q.Y,
	}
}

// Div returns the Vector2D p/k
func (p Vector2D) Div(k float64) Vector2D {
	return Vector2D{
		X: p.X / k,
		Y: p.Y / k,
	}
}

// Eq returns whether the Vector2Ds are approximately equal
func (p Vector2D) Eq(q Vector2D) bool {
	return p.Sub(q).Mag() <= EqThreshold
}

// Mag returns the magnitude (length) of the Vector2D
func (p Vector2D) Mag() float64 {
	return math.Sqrt(math.Pow(p.X, 2) + math.Pow(p.Y, 2))
}

// MagSq returns the magnitude squared (useful for avoiding sqrt calculations)
func (p Vector2D) MagSq() float64 {
	return math.Pow(p.X, 2) + math.Pow(p.Y, 2)
}

// MagManhattan returns the "Manhattan distance" (an approximation)
func (p Vector2D) MagManhattan() float64 {
	return math.Abs(p.X) + math.Abs(p.Y)
}

// Mul returns the Vector2D p*k
func (p Vector2D) Mul(k float64) Vector2D {
	return Vector2D{
		X: p.X * k,
		Y: p.Y * k,
	}
}

// Sub returns the Vector2D p - q
func (p Vector2D) Sub(q Vector2D) Vector2D {
	return Vector2D{
		X: p.X - q.X,
		Y: p.Y - q.Y,
	}
}

func (p Vector2D) Normalize() Vector2D {
	return p.Div(p.Mag())
}

// String returns a string representation of p, like "(1.0,2.0)"
func (p Vector2D) String() string {
	return fmt.Sprintf("(%.3f,%.3f)", p.X, p.Y)
}
