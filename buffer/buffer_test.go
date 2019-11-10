package buffer

import (
	"fmt"
	"testing"
)

func TestBuffer(t *testing.T) {
	buf := NewBuffer(10)
	buf.WriteBytes([]byte{0, 1, 0, 1}, 0)
	buf.WriteBytes([]byte{0, 1, 0, 1, 0, 1, 0, 1}, 4)
	fmt.Println(buf.Bytes())
	fmt.Println(buf.Size())
}

func TestIterator(t *testing.T) {
	buf := make([]byte, 56)
	err := Iterator(buf, 1024, func(idx int, buf []byte) error {
		fmt.Println(buf)
		return nil
	})
	fmt.Println(err)
}
