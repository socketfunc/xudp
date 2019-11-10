package xudp

import (
	"fmt"
	"testing"
)

func TestInitFrame(t *testing.T) {
	frame := &InitFrame{
		StreamID: 1,
		Version:  1,
	}
	fmt.Println(frame.Bytes())
}

func TestInitAckFrame(t *testing.T) {
	frame := &InitAckFrame{
		StreamID: 100,
		Token:    [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
	}
	fmt.Println(frame.Bytes())
}
