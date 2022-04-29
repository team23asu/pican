package demo

import (
	"image/color"
	"math"

	"github.com/team23asu/pican/pkg/can"
	"github.com/team23asu/pican/pkg/rnet"
	"golang.org/x/image/colornames"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func toDegrees(rads float64) float64 {
	return rads * 180.0 / math.Pi
}

func toRads(degrees float64) float64 {
	return math.Pi * degrees / 180.0
}

const (
	// update interval is always 60 Hz or 1/60 seconds
	INPUT_SCALE_SIDE = -1.0
	INPUT_SCALE_FWD  = 1.0

	// METERS_PER_MILE = 1609.34
	// HOURS_PER_SECOND = 1.0 / 60.0 / 60.0
	METERS_PER_INCH  = 0.0254
	MILES_PER_HOUR   = 0.44704 // 1 mph, in meters per second
	PIXELS_PER_METER = 25      // arbitrary

	// page 37, Permobil M300 Corpus HD User Manual:
	// https://permobilwebcdn.azureedge.net/media/yeua1ubg/m300_corpus_hd-user_manual-eng-us-334794.pdf
	CHAIR_WIDTH_METERS           = 0.860
	CHAIR_LENGTH_METERS          = 1.190
	CHAIR_MIN_TURN_RADIUS_METERS = 0.560
	CHAIR_WEIGHT_POUNDS          = 450.0

	CHAIR_SPEED_DAMPING   = 0.7   // arbitrary
	CHAIR_SPEED           = 160.0 // no idea
	CHAIR_ANGULAR_SPEED   = 5.0   // even less of an idea
	CHAIR_ANGULAR_DAMPING = 0.7   // ohh yeah
)

var (
	CHAIR_INDOOR_SPEEDS_M_S = []float64{
		0.8 * MILES_PER_HOUR, // 0.3576 m/s
		1.3 * MILES_PER_HOUR,
		1.7 * MILES_PER_HOUR,
		2.2 * MILES_PER_HOUR,
		2.7 * MILES_PER_HOUR,
	}

	CHAIR_INDOOR_ANGULAR_SPEEDS_RAD_S = []float64{
		0.319,
		0.429,
		0.565,
		0.693,
		0.782,
	}

	CHAIR_MAX_INDOOR_ANGULAR_SPEEDS_RAD_S = []float64{
		CHAIR_INDOOR_SPEEDS_M_S[0] / CHAIR_MIN_TURN_RADIUS_METERS,
		CHAIR_INDOOR_SPEEDS_M_S[1] / CHAIR_MIN_TURN_RADIUS_METERS,
		CHAIR_INDOOR_SPEEDS_M_S[2] / CHAIR_MIN_TURN_RADIUS_METERS,
		CHAIR_INDOOR_SPEEDS_M_S[3] / CHAIR_MIN_TURN_RADIUS_METERS,
		CHAIR_INDOOR_SPEEDS_M_S[4] / CHAIR_MIN_TURN_RADIUS_METERS,
	}

	CHAIR_WIDTH_PIXELS  = CHAIR_WIDTH_METERS * PIXELS_PER_METER
	CHAIR_LENGTH_PIXELS = CHAIR_LENGTH_METERS * PIXELS_PER_METER
)

type Chair struct {
	// speedMode    int // 0, 1. 0=indoor, 1=normal
	speedSetting int // 0 through 4

	// chair position
	position   Vector2D
	bearingDeg float64

	// chair joystick input
	joySide, joyForward float64

	// chair momentum
	momX, momY float64
	// chair health (haha)
	durability int

	bus <-chan *can.Frame // the chair reads movement frames from this channel
	img *ebiten.Image
}

func NewChair(x, y float64, bus chan *can.Frame) *Chair {
	img := generateChairImage(colornames.Blueviolet)

	return &Chair{
		speedSetting: 0,
		bearingDeg:   0.0,
		position:     Vector2D{x, y},
		bus:          bus,
		img:          img,
	}
}

func (c *Chair) SetPosition(x, y float64) {
	c.position = Vector2D{x, y}
}

func (c *Chair) SetBearing(degrees float64) {
	c.bearingDeg = degrees
}

func (c *Chair) SetSpeed(s int) {
	c.speedSetting = s % len(CHAIR_INDOOR_SPEEDS_M_S)
}

func (c *Chair) Update() error {
	// todo: if button pressed, reset position and/or quit

	// // if there is an existing velocity, simulate decay due to friction
	// c.joyForward *= CHAIR_SPEED_DAMPING
	// c.joySide *= CHAIR_ANGULAR_DAMPING
	c.joySide = 0.0
	c.joyForward = 0.0

	// read movement frame off c.bus
	select {
	case f := <-c.bus:
		if rnet.IsMovementFrame(f.ID) {
			c.joyForward, c.joySide = rnet.ConvertDataToJoy(f.Data[0], f.Data[1])
		}
	default:
		//
	}

	fwdLen := c.joyForward * CHAIR_INDOOR_SPEEDS_M_S[c.speedSetting]
	// recall: arc length = radius*theta, so MAX{ theta } = length / MIN { radius }
	// so we restrict the max turning angle to whichever would produce an arc of the forward motion's length
	maxTurnRads := fwdLen / CHAIR_MIN_TURN_RADIUS_METERS
	// desiredRads := CHAIR_INDOOR_ANGULAR_SPEEDS_RAD_S[c.speedSetting] * c.joySide
	// turnRads := math.Min(maxTurnRads, math.Abs(desiredRads))
	// turnRads := maxTurnRads
	// c.bearingDeg += math.Copysign(toDegrees(turnRads), c.joySide)
	c.bearingDeg += toDegrees(c.joySide * maxTurnRads)
	if c.bearingDeg >= 360.0 {
		c.bearingDeg -= 360.0
	}
	if c.bearingDeg < 0.0 {
		c.bearingDeg += 360.0
	}
	bearing := Vector2D{-math.Sin(c.bearingDeg * math.Pi / 180.0), math.Cos(c.bearingDeg * math.Pi / 180.0)}.Normalize()

	c.position = c.position.Add(bearing.Mul(fwdLen * PIXELS_PER_METER))

	return nil
}

func (c *Chair) Draw(screen *ebiten.Image) {
	sw, sh := screen.Size()
	cw, ch := c.img.Size()
	if c.img != nil {
		// op := rotate(c.img, c.bearingDeg)
		// op.GeoM.Translate(c.position.X, c.position.Y)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sw-cw)/2, float64((sh-ch)/2))
		screen.DrawImage(c.img, op)
	}
	// ebitenutil.DebugPrintAt(screen, fmt.Sprintf("pos: (%.1f, %.1f)", c.position.X, c.position.Y), (sw-cw)/2, ch+sh/2)
	ebitenutil.DrawLine(
		screen,
		screenWidth/2,
		screenHeight/2,
		INPUT_SCALE_SIDE*JOY_LINE_LENGTH*c.joySide+screenWidth/2,
		INPUT_SCALE_FWD*JOY_LINE_LENGTH*c.joyForward+screenHeight/2,
		colornames.Yellow,
	)
}

func generateChairImage(clr color.Color) *ebiten.Image {
	img := ebiten.NewImage(int(math.Ceil(CHAIR_WIDTH_PIXELS)), int(math.Ceil(CHAIR_LENGTH_PIXELS)))
	ebitenutil.DrawRect(img, 0, 0, CHAIR_WIDTH_PIXELS, CHAIR_LENGTH_PIXELS, clr)
	// ebitenutil.DrawLine(img, 0, 0, CHAIR_WIDTH_PIXELS, 0, clr)
	// ebitenutil.DrawLine(img, CHAIR_WIDTH_PIXELS, 0, CHAIR_WIDTH_PIXELS, CHAIR_LENGTH_PIXELS, clr)
	// ebitenutil.DrawLine(img, CHAIR_WIDTH_PIXELS, CHAIR_LENGTH_PIXELS, 0, CHAIR_LENGTH_PIXELS, clr)
	// ebitenutil.DrawLine(img, 0, CHAIR_LENGTH_PIXELS, 0, 0, clr)
	// add a small line to indicate front of chair
	ebitenutil.DrawLine(img, CHAIR_WIDTH_PIXELS/2, 0, CHAIR_WIDTH_PIXELS/2, math.Ceil(0.1*PIXELS_PER_METER), colornames.Black)

	return img
}

func rotate(img *ebiten.Image, degrees float64) *ebiten.DrawImageOptions {
	w, h := img.Size()
	op := &ebiten.DrawImageOptions{}

	// Move the image's center to the screen's upper-left corner.
	// This is a preparation for rotating. When geometry matrices are applied,
	// the origin point is the upper-left corner.
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	// Rotate the image. As a result, the anchor point of this rotate is
	// the center of the image.
	op.GeoM.Rotate(toRads(degrees)) // and subtract 90 because we draw facing up but 0 is facing right.

	return op
}
