package main

import (
	"flag"
	"log"

	"github.com/team23asu/pican/pkg/can"
	"github.com/team23asu/pican/pkg/rnet"
)

var (
	joyIface = flag.String("joy", "vcan0", "name of CAN interface to read JSM events from (default: vcan0")
	busIface = flag.String("bus", "vcan1", "name of CAN interface to write JSM events to (default: vcan1")
)

func main() {
	flag.Parse()

	jread, jsend, err := getchannels(*joyIface)
	if err != nil {
		log.Fatalf("failed to bind to %s: %v", *joyIface, err)
	}
	bread, bsend, err := getchannels(*busIface)
	if err != nil {
		log.Fatalf("failed to bind to %s: %v", *busIface, err)
	}

	for {
		select {
		case f := <-jread:
			if rnet.IsMovementFrame(f.ID) {
				// modify. for now, hard code it to BEEF
				copy(f.Data[:], []uint8{0xBE, 0xEF})
			}
			// forward to bus
			bsend <- f
		case f := <-bread:
			// forward to jsm
			jsend <- f
		}
	}
}

func getchannels(iface string) (read chan *can.Frame, send chan *can.Frame, err error) {
	read = make(chan *can.Frame)
	send = make(chan *can.Frame)
	s, err := can.NewSocketBoundTo(iface)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		for {
			f, err := s.Read()
			if err != nil {
				log.Printf("[%s] read error: %v", iface, err)
			}
			read <- f
		}
	}()
	go func() {
		for {
			f := <-send
			err := s.Send(f)
			if err != nil {
				log.Printf("[%s] send error: %v", iface, err)
			}
		}
	}()
	return read, send, nil
}
