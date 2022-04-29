package demo

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/team23asu/pican/pkg/can"
	"github.com/team23asu/pican/pkg/rnet"
)

// this code is meant to demonstrate the behavior of our proposed collision-detection system

type Avoider struct {
	disabled   bool
	jsmRead    <-chan *can.Frame
	chairSend  chan<- *can.Frame
	sensors    []*Sensor
	bearingDeg float64
}

func NewCollisionAvoider(jsmRead, chairSend chan *can.Frame) *Avoider {
	thresholdMeters := 25.0
	frontPushback := 0.5
	sidePushback := 2.0

	return &Avoider{
		jsmRead:   jsmRead,
		chairSend: chairSend,
		sensors: []*Sensor{
			NewSensor(SENSOR_FRONT_CENTER, thresholdMeters, frontPushback),
			NewSensor(SENSOR_FRONT_LEFT, thresholdMeters, sidePushback),
			NewSensor(SENSOR_FRONT_RIGHT, thresholdMeters, sidePushback),
		},
	}
}

func (a *Avoider) Draw(screen *ebiten.Image) {
	if a.disabled {
		return
	}
	for _, s := range a.sensors {
		s.Draw(screen)
	}
}

func (a *Avoider) Update(chairPosition Vector2D, objects []*Object, chairBearingDeg float64) error {
	a.bearingDeg = chairBearingDeg

	for _, s := range a.sensors {
		s.MeasureDistance(chairPosition, objects)
	}
	// asynchronously modify the movement frame
	go func() {
		select {
		case f := <-a.jsmRead:
			a.chairSend <- a.modifyFrame(f)
		// note: bidirectional communication is required in the actual device but is not necessary in our demo,
		//       so we do not bother wiring up jsmSend and chairRead channels.
		default:
			//
		}
	}()
	return nil
}

func (a *Avoider) modifyFrame(f *can.Frame) *can.Frame {
	// don't modify non-movement frames
	if !rnet.IsMovementFrame(f.ID) || a.disabled {
		return f
	}
	// now basically, the goal here is
	// for each sensor, if the input is aligned with the sensor
	// and it shows a reading less than the threshold distance
	// then attenuate the input in that direction
	fwd, side := rnet.ConvertDataToJoy(f.Data[0], f.Data[1])
	inputVect := Vector2D{X: side, Y: fwd}
	// hasChanged := false
	var closest *Sensor
	for _, s := range a.sensors {
		if s.Triggered() {
			if closest == nil || closest.observedMeters > s.observedMeters {
				closest = s
			}
			// inputVect = inputVect.Add(s.Pushback(a.bearingDeg))
			// hasChanged = true
		}
	}
	if closest == nil {
		return f
	}
	inputVect = inputVect.Add(closest.Pushback(a.bearingDeg))

	// avoid jerk by normalizing to the desired magnitude of the user's input
	inputVect = inputVect.Normalize().Mul(Vector2D{X: side, Y: fwd}.Mag())

	x, y := rnet.ConvertJoyToData(float32(inputVect.X), float32(inputVect.Y))
	xxyy := hex.EncodeToString([]byte{uint8(x), uint8(y)})
	line := fmt.Sprintf("02000%s00#%s", JSM_ID, xxyy)
	fn, err := can.FromLog(line)
	if err != nil {
		log.Printf("error re-building frame: %v", err)
	}
	return fn // modified f
}

func (a *Avoider) IsDisabled() bool {
	return a.disabled
}

func (a *Avoider) SetDisabled(disabled bool) {
	a.disabled = disabled
}
