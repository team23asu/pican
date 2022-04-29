package main

import (
	"fmt"
	"log"

	"github.com/team23asu/pican/pkg/can"
	"github.com/team23asu/pican/pkg/rnet"
)

const (
	// TODO: this is not static
	CAN_IFACE = "vcan0"
)

func main() {
	inputs := []float32{0.05, 0.08, 0.1, 0.3, 0.5, 0.8, 0.9, 1.0, 1.2}
	for _, i := range inputs {
		x, y := rnet.ConvertJoyToData(-1.0*i, i)
		fmt.Printf("in: %f, out: %x, %x\n", i, x, y)
	}

	s := can.Socket{}
	s.BindToInterface(CAN_IFACE)

	for _, l := range []string{
		"123#01020304050607",
		"000#0000",
		"02000100#0064",
		"02000100#6400",
		"02000100#649C",
		"02000100#9C64",
	} {
		f, err := can.FromLog(l)
		if err != nil {
			log.Fatalf("failed to build frame: %v", err)
		}
		err = s.Send(f)
		if err != nil {
			log.Fatalf("failed to send %q frame: %v", l, err)
		}
	}

}
