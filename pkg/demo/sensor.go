package demo

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/team23asu/pican/pkg/poly"
	"golang.org/x/image/colornames"
)

type SensorLocation int

const (
	SENSOR_FRONT_LEFT SensorLocation = iota
	SENSOR_FRONT_CENTER
	SENSOR_FRONT_RIGHT
)

type Sensor struct {
	location        SensorLocation
	thresholdMeters float64
	observedMeters  float64
	pushback        float64
}

func NewSensor(location SensorLocation, thresholdMeters, pushback float64) *Sensor {
	return &Sensor{
		location:        location,
		thresholdMeters: thresholdMeters,
		observedMeters:  0,
		pushback:        pushback,
	}
}

func (s *Sensor) Draw(screen *ebiten.Image) {
	pos := Vector2D{X: screenWidth / 2, Y: (screenHeight - CHAIR_LENGTH_PIXELS) / 2}
	// dist := 0.0
	// if s.observedMeters < s.thresholdMeters {
	// 	dist = s.observedMeters
	// }
	ebitenutil.DrawLine(screen, pos.X, pos.Y, math.Ceil(pos.X+s.observedMeters*math.Cos(s.AimRad())), math.Ceil(pos.Y-s.observedMeters*math.Sin(s.AimRad())), colornames.Greenyellow)
}

func (s *Sensor) MeasureDistance(position Vector2D, objects []*Object) {
	// reset sensor
	s.observedMeters = s.thresholdMeters
	// fire a ray from position toward the direction of s.AimRad()
	d := 0.1                                    // step value
	y, x := math.Sincos(s.AimRad())             // direction
	stepVector := Vector2D{X: d * x, Y: -d * y} // step vector
	testPoint := position                       // initial position

	for objDist := d; objDist < s.thresholdMeters; objDist += d {
		for _, o := range objects {
			if o.shape.Contains(poly.Point{X: testPoint.X, Y: testPoint.Y}) {
				log.Printf("%d point hit: %.2f %.2f", s.location, testPoint.X, testPoint.Y)
				if objDist < s.observedMeters {
					// closest object we've seen, store its distance
					s.observedMeters = objDist
				}
				// done with this object
				continue
			}
		}
		// move our test point in the direction of the sensor
		testPoint = testPoint.Add(stepVector)
	}
	if s.observedMeters < s.thresholdMeters {
		// log.Printf("sensor %d: direction: %.1f", s.location, s.AimDeg())
		// log.Printf("sensor %d: sensor ray: (%v)", s.location, testPoint.Sub(position))
		// log.Printf("sensor %d: observed: %.2f", s.location, s.observedMeters)
	}
}

// AimDeg returns the orientation in degrees of a single sensor, based on its location
func (s *Sensor) AimDeg() float64 {
	// this can be confusing due to tranformation to screen coordinates (up is negative Y)
	// (and on the author's consistency when using trigonometry in the various functions.)
	// this all needs to be cleaned up with a whiteboard session!!
	switch s.location {
	case SENSOR_FRONT_LEFT:
		return 135.0
	case SENSOR_FRONT_CENTER:
		return 90.0
	case SENSOR_FRONT_RIGHT:
		return 45.0
	default:
		return 90.0
	}
}

func (s *Sensor) AimRad() float64 {
	return toRads(s.AimDeg())
}

func (s *Sensor) Triggered() bool {
	if s.observedMeters-s.thresholdMeters < 0 {
		return true
	}
	return false
}

// Pushback returns a vector pointing away from a detected obstacle, if any
func (s *Sensor) Pushback(bearingDeg float64) Vector2D {
	if !s.Triggered() {
		return Vector2D{}
	}

	pushAng := -90.0 // default backward
	switch s.location {
	case SENSOR_FRONT_CENTER:
		pushAng = -90.0 // back
	case SENSOR_FRONT_LEFT:
		pushAng = 0.0 // right
	case SENSOR_FRONT_RIGHT:
		pushAng = 180.0 // left
	default:
	}
	pushAng += bearingDeg

	// scaling factor, exponentially ramp with proximity:
	scale := math.Pow(math.Min(1.0, math.Max(1.0-s.observedMeters/s.thresholdMeters, 0.0)), 1.0)
	// scale := 1.0

	return Vector2D{
		X: math.Cos(toRads(pushAng)), Y: math.Sin(toRads(pushAng)),
	}.Mul(s.pushback * scale)
}
