package demo

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/team23asu/pican/pkg/poly"
	"golang.org/x/image/colornames"
)

const (
	// copied from main.go, dont forget to update
	screenWidth  = 640
	screenHeight = 480
)

type Object struct {
	shape *poly.Polygon
	image *ebiten.Image
	imgop *ebiten.DrawImageOptions
	angle float64
}

type World struct {
	chair     *Chair
	avoidance *Avoider
	objects   []*Object
	worldview *ebiten.Image
}

func NewWorld(c *Chair, a *Avoider) *World {
	return &World{
		chair:     c,
		avoidance: a,
		objects:   generateObjects(),
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	sw, sh := screen.Size()
	// we intend to rotate the drawing according to the chair's rotation,
	// so make it sqrt(2) times the dimension of the longest screen side,
	// so that it's guaranteed to fill the screen at any angle of rotation.
	// (also we store it as half the size because it reduces division later on.)
	halfL := math.Ceil(math.Sqrt2 * math.Max(float64(sw)/2, float64(sh)/2))
	w.worldview = ebiten.NewImage(int(2*halfL), int(2*halfL))
	cx, cy := w.chair.position.X, w.chair.position.Y
	for _, o := range w.objects {
		// if !o.shape.BoundingBoxOverlaps(cx-halfL, cy-halfL, cx+halfL, cy+halfL) {
		// 	// object off-screen, skip it
		// 	log.Printf("cam box %.2f %.2f, %.2f %.2f", cx-halfL, cy-halfL, cx+halfL, cy+halfL)
		// 	a, b, c, d := o.shape.BoundingBox()
		// 	log.Printf("obj box %.2f %.2f, %.2f %.2f", a, b, c, d)
		// 	continue
		// }
		if o.image != nil {
			op := &ebiten.DrawImageOptions{}
			ox, oy, _, _ := o.shape.BoundingBox()
			// log.Printf("obj x, y: %.2f, %.2f", ox, oy)
			op.GeoM.Translate(halfL+ox-cx, halfL+oy-cy)
			w.worldview.DrawImage(o.image, op)
		} else {
			log.Printf("obj %v missing image or image options", o)
		}
	}
	// DrawImageOptions for rotating the worldview
	op := &ebiten.DrawImageOptions{}

	// Move the worldview's center to the screen's upper-left corner.
	// This is a preparation for rotating. When geometry matrices are applied,
	// the origin point is the upper-left corner.
	op.GeoM.Translate(-halfL, -halfL)

	// Rotate the image. As a result, the anchor point of this rotate is
	// the center of the image.
	op.GeoM.Rotate(toRads(-w.chair.bearingDeg))

	// Move the center of the rotated worldview to the middle of the screen
	op.GeoM.Translate(screenWidth/2, screenHeight/2)
	screen.DrawImage(w.worldview, op)
	w.chair.Draw(screen)

	w.avoidance.Draw(screen)
}

func (w *World) Update() error {
	err := w.avoidance.Update(w.chair.position.Add(Vector2D{
		X: CHAIR_WIDTH_PIXELS / 2.0, Y: 0, //-CHAIR_LENGTH_PIXELS / 2.0,
	}), w.objects, w.chair.bearingDeg)
	if err != nil {
		log.Printf("collision avoidance error: %v", err)
		return nil
	}

	lastPos := w.chair.position
	err = w.chair.Update()
	if err != nil {
		log.Printf("chair update error: %v", err)
		return nil
	}

	for _, o := range w.objects {
		// do a quick collision check
		for _, p := range []poly.Point{
			{X: w.chair.position.X + 1, Y: w.chair.position.Y + 1 - CHAIR_LENGTH_PIXELS/2},
			{X: w.chair.position.X + CHAIR_WIDTH_PIXELS - 1, Y: w.chair.position.Y + 1 - CHAIR_LENGTH_PIXELS/2},
			{X: w.chair.position.X + CHAIR_WIDTH_PIXELS - 1, Y: w.chair.position.Y - 1 + CHAIR_LENGTH_PIXELS/2},
			{X: w.chair.position.X + 1, Y: w.chair.position.Y - 1 + CHAIR_LENGTH_PIXELS/2},
		} {
			if o.shape.Contains(p) {
				// prevent floating through walls
				w.chair.position = lastPos
			}
		}
		if o.image == nil {
			o.image = o.generateImage(colornames.Aliceblue)
		}
	}
	return nil
}

func (o *Object) generateImage(clr color.Color) *ebiten.Image {
	x1, y1, x2, y2 := o.shape.BoundingBox()

	w, h := math.Ceil(x2-x1), math.Ceil(y2-y1)
	img := ebiten.NewImage(int(w), int(h))
	// bounding box
	ebitenutil.DrawLine(img, 0, 0, w, 0, clr)
	ebitenutil.DrawLine(img, w, 0, w, h, clr)
	ebitenutil.DrawLine(img, w, h, 0, h, clr)
	ebitenutil.DrawLine(img, 0, h, 0, 0, clr)

	// draw edges, relative to bounding box origin of (x1, y1)
	for _, e := range o.shape.Edges() {
		ebitenutil.DrawLine(img, e[0].X-x1, e[0].Y-y1, e[1].X-x1, e[1].Y-y1, colornames.Saddlebrown)
	}

	return img
}

func generateObjects() []*Object {
	objs := []*Object{
		box(100, 100, 100, 100),
	}
	objs = append(objs,
		corridor(-50, -850, 2.5*PIXELS_PER_METER, 700, 1.0*PIXELS_PER_METER)...,
	)
	return objs
}

func box(x, y, w, h float64) *Object {
	return &Object{
		shape: poly.New(
			poly.Point{X: x, Y: y},
			poly.Point{X: x + w, Y: y},
			poly.Point{X: x + w, Y: y + h},
			poly.Point{X: x, Y: y + h},
		),
	}
}

func corridor(x, y, space, l, thickness float64) []*Object {
	return []*Object{
		// left wall
		box(x, y, thickness, l),
		// right wall
		box(x+space+thickness, y, thickness, l),
	}
}
