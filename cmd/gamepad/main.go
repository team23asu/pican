package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type gamepadSet map[ebiten.GamepadID]struct{}

type GUI struct {
	gamepads gamepadSet
}

func NewGUI() *GUI {
	g := &GUI{
		gamepads: make(gamepadSet),
	}
	return g
}

func (g *GUI) Update() error {
	ids := inpututil.AppendJustConnectedGamepadIDs([]ebiten.GamepadID{})
	for _, id := range ids {
		log.Printf("gamepad connected: id %d, SDL: %s", id, ebiten.GamepadSDLID(id))
		g.gamepads[id] = struct{}{}
	}
	return nil
}

func (g *GUI) Draw(screen *ebiten.Image) {
	if len(g.gamepads) <= 0 {
		ebitenutil.DebugPrint(screen, "gamepad not connected")
		return
	}
	for id := range g.gamepads {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			lx := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
			ly := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
			rx := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickHorizontal)
			ry := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickVertical)
			ebitenutil.DebugPrint(
				screen,
				fmt.Sprintf(
					"standard gamepad:\n  Left stick: <X %+0.2f, Y %+0.2f>\n  Right stick: <X %+0.2f, Y %+0.2f>",
					lx, ly,
					rx, ry,
				),
			)
		}
	}
}

func (g *GUI) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := NewGUI()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Team 23 Gamepad Util")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
