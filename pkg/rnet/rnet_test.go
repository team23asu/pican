package rnet

import (
	"testing"

	"golang.org/x/sys/unix"
)

// func TestSplit(t *testing.T) {
// 	type test struct {
// 			input string
// 			sep   string
// 			want  []string
// 	}

// 	tests := []test{
// 			{input: "a/b/c", sep: "/", want: []string{"a", "b", "c"}},
// 			{input: "a/b/c", sep: ",", want: []string{"a/b/c"}},
// 			{input: "abc", sep: "/", want: []string{"abc"}},
// 	}

// 	for _, tc := range tests {
// 			got := Split(tc.input, tc.sep)
// 			if !reflect.DeepEqual(tc.want, got) {
// 					t.Fatalf("expected: %v, got: %v", tc.want, got)
// 			}
// 	}
// }

// func TestIsMovementFrame(t *testing.T) {
// 	id := uint32(0x02000700)

// }

func TestIsMovementFrame(t *testing.T) {
	type test struct {
		id   uint32
		want bool
	}

	tests := []test{
		{id: 0x02000100, want: false},
		{id: 0x02000100 | unix.CAN_EFF_FLAG, want: true},
		{id: 0x02000200 | unix.CAN_EFF_FLAG, want: true},
		{id: 0x02000E00 | unix.CAN_EFF_FLAG, want: true},
		{id: 0x02000F00 | unix.CAN_EFF_FLAG, want: true},
		{id: 0x02000101 | unix.CAN_EFF_FLAG, want: false},
		{id: 0x02000101 | unix.CAN_EFF_FLAG, want: false},
		{id: 0xFDFFF0FF | unix.CAN_EFF_FLAG, want: false},
	}

	for _, test := range tests {
		got := IsMovementFrame(test.id)
		if got != test.want {
			t.Fatalf("IsMovementFrame(0x%.8x), expected: %t, got: %v", test.id, test.want, got)
		}
	}
}
