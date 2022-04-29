package demo

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"

	"github.com/team23asu/pican/pkg/can"
	"github.com/team23asu/pican/pkg/rnet"
)

const (
	FRAME_INTERVAL = 500 * time.Millisecond // 10 millisecond is actual R-Net value but we may not be able to update this fast given that we're drawing the screen too
	JSM_ID         = "1"

	JOY_LINE_LENGTH = 100
)

// applies a curve to the input, i.e. convert from a linear function to logarithmic
// this is meant to be used with normalized (0.0 to 1.0) input
func applyCurve(value, exp, scale float64) float64 {
	return scale * math.Copysign(math.Pow(math.Abs(value), exp), value)
}

type GamepadSet struct {
	set            map[ebiten.GamepadID]struct{}
	lx, ly, rx, ry float64
	bus            chan<- *can.Frame // the gamepad emits movement frames onto this channel to be read elsewhere

	mostRecentFrame *can.Frame // just used for drawing on screen.
}

func NewGamepadSet(bus chan *can.Frame) *GamepadSet {
	return &GamepadSet{
		set: make(map[ebiten.GamepadID]struct{}),
		bus: bus,
	}
}

func (g *GamepadSet) Draw(screen *ebiten.Image) {
	if len(g.set) <= 0 {
		ebitenutil.DebugPrint(screen, "gamepad not connected")
		// return
	}
	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf(
			"\n  Left stick: <X %+0.2f, Y %+0.2f>\n  Right stick: <X %+0.2f, Y %+0.2f>",
			g.lx, g.ly,
			g.rx, g.ry,
		),
	)

	if g.mostRecentFrame != nil {
		ebitenutil.DebugPrintAt(
			screen,
			fmt.Sprintf("R-Net payload: #%s", hex.EncodeToString(g.mostRecentFrame.Data[:g.mostRecentFrame.DLC])),
			screenWidth/2-50,
			screenHeight/2+48,
		)
	}
	ebitenutil.DrawLine(
		screen,
		screenWidth/2,
		screenHeight/2,
		JOY_LINE_LENGTH*g.lx+screenWidth/2,
		JOY_LINE_LENGTH*g.ly+screenHeight/2,
		colornames.Antiquewhite,
	)
}

func (g *GamepadSet) GetButton(b ebiten.StandardGamepadButton) bool {
	for id := range g.set {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			return ebiten.IsStandardGamepadButtonPressed(id, b)
		}
	}
	return false
}

func (g *GamepadSet) Update() error {
	g.lx, g.ly, g.rx, g.ry = 0, 0, 0, 0
	ids := inpututil.AppendJustConnectedGamepadIDs([]ebiten.GamepadID{})
	for _, id := range ids {
		log.Printf("gamepad connected: id %d, SDL: %s", id, ebiten.GamepadSDLID(id))
		g.set[id] = struct{}{}
	}
	for id := range g.set {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			g.lx = applyCurve(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal), 1.6, 0.5)
			g.ly = applyCurve(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical), 1.6, 0.75)
			g.rx = applyCurve(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickHorizontal), 1.0, 1.0)
			g.ry = applyCurve(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickVertical), 1.0, 1.0)
			// only use first gamepad found
			break
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.ly = -0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.ly = 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.lx = -0.25
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.lx = 0.25
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.ry = -0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.ry = 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.rx = -0.25
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.rx = 0.25
	}
	x, y := rnet.ConvertJoyToData(float32(g.lx), float32(g.ly))
	xxyy := hex.EncodeToString([]byte{uint8(x), uint8(y)})
	// note: +100 = 0x64 and -100 = 0x9C (two's complement of 0x64)
	line := fmt.Sprintf("02000%s00#%s", JSM_ID, xxyy)
	f, err := can.FromLog(line)
	if err != nil {
		log.Printf("error building frame: %v", err)
	}
	select {
	case g.bus <- f:
		g.mostRecentFrame = f
		// log.Printf("[sent] %s", line)
	default:
		// log.Printf("[drop] %s", line)
	}
	return nil
}
