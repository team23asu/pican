package can

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// The basic CAN frame structure and the sockaddr structure are defined
// in include/linux/can.h:

//   struct can_frame {
//           canid_t can_id;  /* 32 bit CAN_ID + EFF/RTR/ERR flags */
//           __u8    can_dlc; /* frame payload length in byte (0 .. 8) */
//           __u8    __pad;   /* padding */
//           __u8    __res0;  /* reserved / padding */
//           __u8    __res1;  /* reserved / padding */
//           __u8    data[8] __attribute__((aligned(8)));
//   };
//
// from https://www.kernel.org/doc/Documentation/networking/can.txt
type Frame struct {
	ID   uint32   // 4 byte: either 11-bit (3 hex chars) or 29-bit (8 hex chars)
	DLC  uint8    // 1 byte: data length
	Pad  uint8    // 1 byte
	Res0 uint8    // 1 byte
	Res1 uint8    // 1 byte
	Data [8]uint8 // 8 byte: 16 hex chars (uint8 = 2 hex chars = one byte)
}

const (
	FRAME_MAX_SIZE int = 16 // 16-byte maximum frame size

	CAN_EFF_FLAG = 0x80000000
)

func (f Frame) Payload() []byte {
	return f.Data[:f.DLC]
}

// use candump from can-utils log format
// Usage: cansend <device> <can_frame>.

// <can_frame>:
//  <can_id>#{data}          for 'classic' CAN 2.0 data frames
//  <can_id>#R{len}          for 'classic' CAN 2.0 data frames
//  <can_id>##<flags>{data}  for CAN FD frames

// <can_id>:
//  3 (SFF) or 8 (EFF) hex chars
// {data}:
//  0..8 (0..64 CAN FD) ASCII hex-values (optionally separated by '.')
// {len}:
//  an optional 0..8 value as RTR frames can contain a valid dlc field
// <flags>:
//  a single ASCII Hex value (0 .. F) which defines canfd_frame.flags
func FromLog(logline string) (*Frame, error) {
	if strings.Contains(logline, "R") {
		return nil, fmt.Errorf("invalid input: 'R' not supported")
	}
	parts := strings.Split(logline, "#")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid input: expected single # separator")
	}
	if len(parts[0]) != 3 && len(parts[0]) != 8 {
		return nil, fmt.Errorf("invalid input: id expected to be 3 or 8 hex chars")
	}
	if len(parts[1])%2 != 0 {
		return nil, fmt.Errorf("invalid input: data expected to be even-numbered hex chars")
	}
	if len(parts[1]) > 16 {
		return nil, fmt.Errorf("invalid input: max data length is 16 hex chars")
	}
	if len(parts[1]) < 2 {
		return nil, fmt.Errorf("invalid input: missing data")
	}
	isExtendedID := len(parts[0]) == 8
	if len(parts[0]) == 3 {
		// make id 8 hex chars long for convenience
		parts[0] = "00000" + parts[0]
	}
	idBytes, err := hex.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid input: id not hex-encoded uint32")
	}
	id := binary.BigEndian.Uint32(idBytes)
	if isExtendedID {
		id = id | CAN_EFF_FLAG
	}

	dataBytes, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid input: data not hex-encoded []uint8")
	}
	f := &Frame{
		ID:  id,
		DLC: uint8(len(dataBytes)),
	}
	copy(f.Data[:], dataBytes)
	return f, nil
}
