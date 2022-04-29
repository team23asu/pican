// go:build linux

package can

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/sys/unix"
)

type Socket struct {
	fd int
}

func NewSocketBoundTo(iface string) (*Socket, error) {
	s := &Socket{}
	err := s.BindToInterface(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to %q: %w", iface, err)
	}
	return s, nil
}

func (s *Socket) BindToInterface(name string) error {
	i, err := net.InterfaceByName(name)
	if err != nil {
		log.Fatalf("failed to get %q: %v", name, err)
	}

	fd, err := unix.Socket(unix.AF_CAN, unix.SOCK_RAW, unix.CAN_RAW)
	if err != nil {
		log.Fatalf("failed to get socket: %v", err)
	}

	// see also https://pkg.go.dev/golang.org/x/sys/unix#SockaddrCAN
	sa := &unix.SockaddrCAN{Ifindex: i.Index}
	if err := unix.Bind(fd, sa); err != nil {
		log.Fatalf("failed to bind socket: %v", err)
	}
	s.fd = fd
	return nil
}

func (s *Socket) Send(f *Frame) error {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, f)
	if err != nil {
		return fmt.Errorf("unable to encode frame: %w", err)
	}
	if buf.Len() > int(FRAME_MAX_SIZE) {
		return fmt.Errorf("frame too big. 16-byte maximum")
	}
	n, err := unix.Write(s.fd, buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}
	log.Printf("wrote %d bytes to fd %d", n, s.fd)
	return nil
}

func (s *Socket) Read() (*Frame, error) {
	buf := make([]byte, FRAME_MAX_SIZE)
	n, err := unix.Read(s.fd, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read %d bytes: %w", n, err)
	}
	msg := &Frame{}
	err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, msg)
	if err != nil {
		return nil, fmt.Errorf("unable to decode frame: %w", err)
	}
	log.Printf("read frame from %d", s.fd)
	return msg, nil
}
