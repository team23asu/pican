package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/team23asu/pican/pkg/can"
	"github.com/team23asu/pican/pkg/demo"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type gamepadSet map[ebiten.GamepadID]struct{}

type Demo struct {
	gamepads  *demo.GamepadSet
	chair     *demo.Chair
	avoidance *demo.Avoider
	world     *demo.World
}

func NewDemo() *Demo {
	// create channels used for passing can frames from JSM to chair via our hardware.
	jsmRead := make(chan *can.Frame, 1)
	chairSend := make(chan *can.Frame, 1)
	// jsmSend := make(chan *can.Frame, 1) // needed in hardware but not needed in demo
	// chairRead := make(chan *can.Frame, 1) // needed in hardward but not needed in demo

	// plug one end of a cable into the virtual chair
	c := demo.NewChair(0, 0, chairSend)

	// plug one end of a different cable into the virtual JSM
	g := demo.NewGamepadSet(jsmRead)

	// plug collision module in between the chair and JSM
	a := demo.NewCollisionAvoider(jsmRead, chairSend)

	return &Demo{
		gamepads:  g,
		chair:     c,
		avoidance: a,
		world:     demo.NewWorld(c, a),
	}
}

func (d *Demo) Draw(screen *ebiten.Image) {
	// d.chair.Draw(screen)
	d.world.Draw(screen)
	d.gamepads.Draw(screen)
	msg := "CONTROLS:\n"
	msg += "MOVE: left stick\n"
	if d.avoidance.IsDisabled() {
		msg += "ENABLE AVOIDANCE: key Z (it is DISABLED!)\n"
	} else {
		msg += "DISABLE AVOIDANCE: key Z\n"
	}
	msg += "RESET: right-most button\n"
	msg += "SPEED: keys 1-5\n"
	msg += "QUIT: select (center-left) or ESC\n"
	ebitenutil.DebugPrintAt(screen, msg, 10, 100)
}

func (d *Demo) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || d.gamepads.GetButton(ebiten.StandardGamepadButtonCenterLeft) {
		log.Printf("key ESC pressed, goodbye!")
		os.Exit(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		d.avoidance.SetDisabled(!d.avoidance.IsDisabled())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyB) || d.gamepads.GetButton(ebiten.StandardGamepadButtonRightRight) {
		d.chair.SetPosition(0, 0)
		// g.chair.SetBearing(0.0)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		d.chair.SetSpeed(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		d.chair.SetSpeed(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		d.chair.SetSpeed(2)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		d.chair.SetSpeed(3)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		d.chair.SetSpeed(4)
	}

	err := d.gamepads.Update()
	if err != nil {
		return err
	}

	err = d.world.Update()
	if err != nil {
		return err
	}
	return nil
}

func (d *Demo) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	d := NewDemo()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Team 23 Demo")
	// ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(d); err != nil {
		log.Fatal(err)
	}
}
