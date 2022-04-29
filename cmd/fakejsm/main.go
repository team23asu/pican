package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/0xcafed00d/joystick"
	"github.com/team23asu/pican/pkg/can"
	"github.com/team23asu/pican/pkg/rnet"
)

var (
	iface = flag.String("iface", "vcan0", "name of CAN interface to send JSM events to (default: vcan0)")
	jsmid = flag.String("jsmid", "1", "id of R-Net JSM (default: 1)")
	joyid = flag.Int("joyid", 0, "id of USB joystick device (default: 0)")
)

const (
	interval = 1000 * time.Millisecond // 10 millisecond is actual value
)

func main() {
	flag.Parse()
	socket, err := can.NewSocketBoundTo(*iface)
	if err != nil {
		log.Fatal(err)
	}

	js, err := joystick.Open(*joyid)
	if err != nil {
		log.Fatalf("failed to open /dev/input/js%d: %v", *joyid, err)
	}
	defer js.Close()
	log.Printf("Joystick name: %s", js.Name())

	// loop forever
	for {
		state, err := js.Read()
		if err != nil {
			log.Printf("error reading joystick: %v", err)
		}

		lx, ly := state.AxisData[0], state.AxisData[1]
		// rx, ry := state.AxisData[3], state.AxisData[4]

		/* joystick direction and polarity of measurement (up is -Y, left is -X):

		    -Y
		-X      +X
		    +Y

		*/

		x, y := rnet.ConvertJoyToData(float32(lx)/32767.0, float32(ly)/32767.0)
		xxyy := hex.EncodeToString([]byte{uint8(x), uint8(y)})
		// note: +100 = 0x64 and -100 = 0x9C (two's complement of 0x64)
		line := fmt.Sprintf("02000%s00#%s", *jsmid, xxyy)
		// line := fmt.Sprintf("02000%s01#%s", *jsmid, xxyy)
		log.Printf("line: %q", line)
		f, err := can.FromLog(line)
		log.Printf("frame: %v", f)
		log.Printf("data: %v", f.Payload())
		if err != nil {
			log.Printf("error building frame: %v", err)
		}
		err = socket.Send(f)
		if err != nil {
			log.Printf("error sending frame: %v", err)
		}
		time.Sleep(interval)
	}
}
