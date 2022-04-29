// go:build !linux

package can

// Socket is a virtual (in-memory) socket since we do not have the vcan kernel module available
type Socket struct {
	data chan *Frame
}

func NewSocketBoundTo(iface string) (*Socket, error) {
	return &Socket{
		data: make(chan *Frame),
	}, nil
}

func (s *Socket) Send(f *Frame) error {
	s.data <- f
	return nil
}

func (s *Socket) Read() (*Frame, error) {
	f := <-s.data
	return f, nil
}
