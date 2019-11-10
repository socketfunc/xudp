package xudp

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

type ConnectionID [16]byte

type Type uint8

const (
	None Type = iota
	Init
	InitAck
	Session
	SessAck
	Data
	DataAck
	Ping
	Pong
	Shutdown
	ShutAck
)

func (t Type) String() string {
	switch t {
	case Init:
		return "init"
	case InitAck:
		return "initack"
	case Session:
		return "session"
	case SessAck:
		return "sessack"
	case Data:
		return "data"
	case DataAck:
		return "dataack"
	case Ping:
		return "ping"
	case Pong:
		return "pong"
	case Shutdown:
		return "shutdown"
	case ShutAck:
		return "shutack"
	}
	return ""
}

type Packet struct {
	Header PacketHeader
	Frame  Frame
}

type PacketHeader struct {
	checksum     uint32
	Type         Type
	ConnectionID ConnectionID
	Sequence     uint32
	Channel      uint8
}

func (h *PacketHeader) Size() int {
	return 26
}

func (h *PacketHeader) Bytes() []byte {
	buf := make([]byte, 22)
	buf[0] = uint8(h.Type)
	copy(buf[1:17], h.ConnectionID[:])
	binary.BigEndian.PutUint32(buf[17:21], h.Sequence)
	buf[21] = h.Channel
	return buf
}

func NewPacketHeader(typ Type, id []byte, sequence uint32, channel uint8) *PacketHeader {
	h := &PacketHeader{}
	h.Type = typ
	copy(h.ConnectionID[:], id)
	h.Sequence = sequence
	h.Channel = channel
	return h
}

func Payload(h *PacketHeader, data []byte) []byte {
	p := append(h.Bytes(), data...)
	c := make([]byte, 4)
	binary.BigEndian.PutUint32(c, checksum(p))
	return append(c, p...)
}

func NewPacket(id []byte, sequence uint32, channel uint8, frame Frame) *Packet {
	p := &Packet{}
	copy(p.Header.ConnectionID[:], id)
	p.Header.Sequence = sequence
	p.Header.Channel = channel
	p.Frame = frame
	return p
}

func (p *Packet) ConnectionID() ConnectionID {
	return p.Header.ConnectionID
}

func (p *Packet) Channel() uint8 {
	return p.Header.Channel
}

func (p *Packet) Sequence() uint32 {
	return p.Header.Sequence
}

func (p *Packet) NextSequence() uint32 {
	return p.Header.Sequence + 1
}

func (p *Packet) Bytes() []byte {
	frame := p.Frame.Bytes()
	buf := make([]byte, len(frame)+26)
	buf[4] = uint8(p.Frame.Type())
	copy(buf[5:21], p.Header.ConnectionID[:])
	binary.BigEndian.PutUint32(buf[21:25], p.Header.Sequence)
	buf[25] = p.Header.Channel
	copy(buf[26:], frame)
	binary.BigEndian.PutUint32(buf[0:4], checksum(buf[4:]))
	return buf
}

func checksum(buf []byte) uint32 {
	return crc32.Checksum(buf, crc32.IEEETable)
}

func CheckPacket(buf []byte) error {
	c := binary.BigEndian.Uint32(buf[0:4])
	if c != checksum(buf[4:]) {
		return errors.New("packet: checksum error")
	}
	return nil
}

func DecodePacketHeader(buf []byte) (*PacketHeader, error) {
	h := &PacketHeader{}
	h.checksum = binary.BigEndian.Uint32(buf[0:4])
	if h.checksum != checksum(buf[4:]) {
		//return nil, errors.New("packet: checksum error")
	}
	h.Type = Type(buf[4])
	copy(h.ConnectionID[:], buf[5:21])
	h.Sequence = binary.BigEndian.Uint32(buf[21:25])
	h.Channel = buf[25]
	return h, nil
}
