package xudp

import (
	"fmt"
	"testing"
)

func TestDecodePacket(t *testing.T) {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
	f := &InitFrame{
		StreamID: 1,
		Version:  1,
	}
	p := NewPacket(id, 1, 1, f)
	fmt.Println(p.Bytes())

	err := CheckPacket(p.Bytes())
	fmt.Println(err)

	header, err := DecodePacketHeader(p.Bytes())
	fmt.Println(err)
	fmt.Println(header)

	fmt.Println(p.Bytes()[header.Size():])
}

func BenchmarkPacket_Bytes(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
		f := &InitFrame{
			StreamID: 1,
			Version:  1,
		}
		p := NewPacket(id, 1, 1, f)
		p.Bytes()
	}
}
