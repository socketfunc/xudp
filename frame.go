package xudp

import (
	"crypto/rand"
	"encoding/binary"
)

type Frame interface {
	Type() Type
	Bytes() []byte
}

var (
	_ Frame = (*InitFrame)(nil)
	_ Frame = (*InitAckFrame)(nil)
	_ Frame = (*SessionFrame)(nil)
	_ Frame = (*SessAckFrame)(nil)
	_ Frame = (*DataFrame)(nil)
	_ Frame = (*DataAckFrame)(nil)
	_ Frame = (*PingFrame)(nil)
	_ Frame = (*PongFrame)(nil)
	_ Frame = (*ShutdownFrame)(nil)
	_ Frame = (*ShutAckFrame)(nil)
)

func DecodeFrame(typ Type, buf []byte) Frame {
	switch typ {
	case Init:
		return decodeInitFrame(buf)
	case InitAck:
		return decodeInitAckFrame(buf)
	case Session:
		return decodeSessionFrame(buf)
	case SessAck:
		return decodeSessAckFrame(buf)
	case Data:
		return decodeDataFrame(buf)
	case DataAck:
		return decodeDataAckFrame(buf)
	case Ping:
		return decodePingFrame(buf)
	case Pong:
		return decodePongFrame(buf)
	case Shutdown:
		return decodeShutdownFrame(buf)
	case ShutAck:
		return decodeShutAckFrame(buf)
	}
	return nil
}

type InitFrame struct {
	StreamID uint32
	Version  uint32
}

func (f *InitFrame) Type() Type {
	return Init
}

func (f *InitFrame) Bytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	binary.BigEndian.PutUint32(buf[4:8], f.Version)
	return buf
}

func decodeInitFrame(buf []byte) *InitFrame {
	frame := &InitFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	frame.Version = binary.BigEndian.Uint32(buf[4:8])
	return frame
}

type InitAckFrame struct {
	StreamID uint32
	Token    [16]byte
}

func (f *InitAckFrame) Type() Type {
	return InitAck
}

func (f *InitAckFrame) Bytes() []byte {
	buf := make([]byte, 20)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	copy(buf[4:20], f.Token[:])
	return buf
}

func decodeInitAckFrame(buf []byte) *InitAckFrame {
	frame := &InitAckFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	copy(frame.Token[:], buf[4:20])
	return frame
}

type SessionFrame struct {
	StreamID uint32
	Token    [16]byte
	Key      [64]byte
}

func (f *SessionFrame) Type() Type {
	return Session
}

func (f *SessionFrame) Bytes() []byte {
	buf := make([]byte, 84)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	copy(buf[4:20], f.Token[:])
	copy(buf[20:84], f.Key[:])
	return buf
}

func decodeSessionFrame(buf []byte) *SessionFrame {
	frame := &SessionFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	copy(frame.Token[:], buf[4:20])
	copy(frame.Key[:], buf[20:84])
	return frame
}

type SessAckFrame struct {
	StreamID uint32
	Key      [64]byte
}

func (f *SessAckFrame) setKey(key []byte) {
	copy(f.Key[:], key)
}

func (f *SessAckFrame) Type() Type {
	return SessAck
}

func (f *SessAckFrame) Bytes() []byte {
	buf := make([]byte, 68)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	copy(buf[4:68], f.Key[:])
	return buf
}

func decodeSessAckFrame(buf []byte) *SessAckFrame {
	frame := &SessAckFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	copy(frame.Key[:], buf[4:68])
	return frame
}

type DataFrame struct {
	StreamID uint32     // 4
	Offset   int        // 4
	Length   int        // 4
	Hash     [32]byte   // 32
	Size     uint16     // 2
	Data     [1024]byte // 1024
}

func (f *DataFrame) Type() Type {
	return Data
}

func (f *DataFrame) SetData(d []byte) {
	f.Size = uint16(len(d))
	copy(f.Data[:], d)
}

func (f *DataFrame) RawData() ([]byte, error) {
	raw := f.Data[0:f.Size]
	return raw, nil
}

func (f *DataFrame) Bytes() []byte {
	buf := make([]byte, 1070)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	binary.BigEndian.PutUint32(buf[4:8], uint32(f.Offset))
	binary.BigEndian.PutUint32(buf[8:12], uint32(f.Length))
	copy(buf[12:44], f.Hash[:])
	binary.BigEndian.PutUint16(buf[44:46], f.Size)
	copy(buf[46:46+f.Size], f.Data[:])
	if f.Size < 1024 {
		_, _ = rand.Read(buf[46+f.Size:])
	}
	return buf
}

func decodeDataFrame(buf []byte) *DataFrame {
	frame := &DataFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	frame.Offset = int(binary.BigEndian.Uint32(buf[4:8]))
	frame.Length = int(binary.BigEndian.Uint32(buf[8:12]))
	copy(frame.Hash[:], buf[12:44])
	frame.Size = binary.BigEndian.Uint16(buf[44:46])
	copy(frame.Data[:], buf[46:1070])
	return frame
}

type DataAckFrame struct {
	StreamID uint32
	Offset   int
}

func (f *DataAckFrame) Type() Type {
	return DataAck
}

func (f *DataAckFrame) Bytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	binary.BigEndian.PutUint32(buf[4:8], uint32(f.Offset))
	return buf
}

func decodeDataAckFrame(buf []byte) *DataAckFrame {
	frame := &DataAckFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	frame.Offset = int(binary.BigEndian.Uint32(buf[4:8]))
	return frame
}

type PingFrame struct {
	StreamID uint32
}

func (f *PingFrame) Type() Type {
	return Ping
}

func (f *PingFrame) Bytes() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	return buf
}

func decodePingFrame(buf []byte) *PingFrame {
	frame := &PingFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	return frame
}

type PongFrame struct {
	StreamID uint32
}

func (f *PongFrame) Type() Type {
	return Pong
}

func (f *PongFrame) Bytes() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	return buf
}

func decodePongFrame(buf []byte) *PongFrame {
	frame := &PongFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	return frame
}

type ShutdownFrame struct {
	StreamID uint32
}

func (f *ShutdownFrame) Type() Type {
	return Shutdown
}

func (f *ShutdownFrame) Bytes() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	return buf
}

func decodeShutdownFrame(buf []byte) *ShutdownFrame {
	frame := &ShutdownFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	return frame
}

type ShutAckFrame struct {
	StreamID uint32
}

func (f *ShutAckFrame) Type() Type {
	return ShutAck
}

func (f *ShutAckFrame) Bytes() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], f.StreamID)
	return buf
}

func decodeShutAckFrame(buf []byte) *ShutAckFrame {
	frame := &ShutAckFrame{}
	frame.StreamID = binary.BigEndian.Uint32(buf[0:4])
	return frame
}
