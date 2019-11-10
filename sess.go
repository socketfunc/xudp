package xudp

import (
	"crypto/sha256"
	"math/rand"
	"net"
	"time"

	"github.com/DataDog/zstd"
	"github.com/socketfunc/xudp/buffer"
	"github.com/socketfunc/xudp/crypto"
	"golang.org/x/sync/errgroup"
)

const (
	protocol = "udp"
)

type Sess struct {
	dialer bool

	addr        net.Addr
	conn        *net.UDPConn
	incoming    chan []byte
	acknowledge chan []byte
	quit        chan struct{}
	ticker      *time.Ticker

	private   *crypto.PrivateKey
	public    *crypto.PublicKey
	secretKey []byte

	ConnectionID ConnectionID
	Sequence     uint32
	tempData     map[uint32]*buffer.Buffer
	data         map[uint32]*buffer.Buffer
}

func NewSess(conn *net.UDPConn, addr net.Addr, secret []byte) *Sess {
	sess := &Sess{
		addr:        addr,
		conn:        conn,
		incoming:    make(chan []byte, queueSize),
		acknowledge: make(chan []byte, queueSize),
		quit:        make(chan struct{}, 1),
		secretKey:   secret,
		Sequence:    rand.Uint32(),
		data:        map[uint32]*buffer.Buffer{},
	}
	return sess
}

func (s *Sess) keepalive() {
}

func (s *Sess) setConnectionID(id []byte) {
	copy(s.ConnectionID[:], id)
}

func (s *Sess) encryptData(buf []byte) ([]byte, error) {
	return crypto.Encrypt(s.secretKey, buf)
}

func (s *Sess) decryptData(buf []byte) ([]byte, error) {
	return crypto.Decrypt(s.secretKey, buf)
}

func (s *Sess) compressData(buf []byte) ([]byte, error) {
	return zstd.CompressLevel(nil, buf, 9)
}

func (s *Sess) decompressData(buf []byte) ([]byte, error) {
	return zstd.Decompress(nil, buf)
}

func (s *Sess) marshalData(buf []byte) ([]byte, error) {
	buf, err := s.compressData(buf)
	if err != nil {
		return nil, err
	}
	return s.encryptData(buf)
}

func (s *Sess) unmarshalData(buf []byte) ([]byte, error) {
	buf, err := s.decryptData(buf)
	if err != nil {
		return nil, err
	}
	return s.decompressData(buf)
}

func (s *Sess) read() ([]byte, error) {
	buf := make([]byte, bufferSize)
	n, err := s.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (s *Sess) send(buf []byte) error {
	if s.dialer {
		_, err := s.conn.Write(buf)
		return err
	}
	_, err := s.conn.WriteTo(buf, s.addr)
	return err
}

func (s *Sess) TempData(frame *DataFrame) {
	if _, ok := s.tempData[frame.StreamID]; !ok {
		s.tempData[frame.StreamID] = buffer.NewBuffer(frame.Length)
	}
}

func (s *Sess) Close() error {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	return nil
}

func (s *Sess) NextSequence() uint32 {
	s.Sequence++
	return s.Sequence
}

func (s *Sess) RemoteAddr() string {
	return s.addr.String()
}

func (s *Sess) Receive() ([]byte, error) {
	select {
	case in := <-s.incoming:
		return in, nil
	}
}

func (s *Sess) Send(buf []byte) error {
	const chunkSize = 1024
	eg := errgroup.Group{}
	eg.Go(func() error {
		streamID := rand.Uint32()
		hash := sha256.Sum256(buf)
		length := len(buf)
		err := buffer.Iterator(buf, chunkSize, func(offset int, chunk []byte) error {
			f := &DataFrame{
				StreamID: streamID,
				Length:   length,
				Offset:   offset,
				Hash:     hash,
			}
			f.SetData(chunk)
			data, err := s.marshalData(f.Bytes())
			if err != nil {
				return err
			}
			header := &PacketHeader{
				Type:         Data,
				ConnectionID: s.ConnectionID,
				Sequence:     s.NextSequence(),
				Channel:      1,
			}
			return s.send(Payload(header, data))
		})
		return err
	})
	return eg.Wait()
}
