package xudp

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/socketfunc/xudp/buffer"
	"github.com/socketfunc/xudp/crypto"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	bufferSize = 4096
	queueSize  = 128
)

type Conn struct {
	conn      *net.UDPConn
	sessions  sync.Map
	bytePool  sync.Pool
	accepting chan *Sess
	quit      chan struct{}
}

func (c *Conn) listen() {
	go func() {
		for {
			if err := c.loop(); err != nil {
				log.Println(err)
			}
			select {
			case <-c.quit:
				return
			default:
			}
		}
	}()
}

func (c *Conn) loop() error {
	buf := c.getBufferPool()
	defer c.putBufferPool(buf)
	n, addr, err := c.conn.ReadFrom(buf)
	if err != nil {
		return err
	}
	h, err := DecodePacketHeader(buf[:n])
	if err != nil {
		return err
	}
	if h.Type == None {
		return nil
	}
	switch h.Type {
	case Init:
		return c.initHandler(h, addr)
	case Session:
		frame := decodeSessionFrame(buf[h.Size():n])
		sess, err := c.sessionHandler(h, frame, addr)
		if err != nil {
			return err
		}
		c.setSess(h.ConnectionID, sess)
		c.accepting <- sess
		return nil
	}
	sess, ok := c.getSess(h.ConnectionID)
	if !ok {
		return nil
	}
	data, err := sess.unmarshalData(buf[h.Size():n])
	if err != nil {
		return fmt.Errorf("xudp: decrypt data error. %w", err)
	}
	sess.addr = addr
	switch h.Type {
	case Data:
		frame := decodeDataFrame(data)
		return c.dataHandler(sess, frame)
	case DataAck:
		//frame := decodeDataAckFrame(data)
		//sess.acknowledge <- frame.Bytes()
	}
	return nil
}

func (c *Conn) setSess(id ConnectionID, sess *Sess) {
	c.sessions.Store(id, sess)
}

func (c *Conn) getSess(id ConnectionID) (*Sess, bool) {
	sess, ok := c.sessions.Load(id)
	if !ok {
		return nil, ok
	}
	return sess.(*Sess), ok
}

func (c *Conn) getBufferPool() []byte {
	return c.bytePool.Get().([]byte)
}

func (c *Conn) putBufferPool(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
	c.bytePool.Put(buf)
}

func (c *Conn) initHandler(h *PacketHeader, addr net.Addr) error {
	f := &InitAckFrame{
		StreamID: rand.Uint32(),
		Token:    c.createToken(addr),
	}
	ack := NewPacket(h.ConnectionID[:], h.Sequence+1, h.Channel, f)
	_, err := c.conn.WriteTo(ack.Bytes(), addr)
	return err
}

func (c *Conn) sessionHandler(h *PacketHeader, frame *SessionFrame, addr net.Addr) (*Sess, error) {
	if !c.verifyToken(frame.Token) {
		return nil, errors.New("xudp: invalid token")
	}
	private, public, err := crypto.GenerateKeys()
	if err != nil {
		return nil, err
	}
	pk := crypto.GeneratePublicKey(frame.Key[:])
	secret, err := crypto.ComputeSecret(private, pk)
	if err != nil {
		return nil, err
	}
	f := &SessAckFrame{
		StreamID: rand.Uint32(),
	}
	f.setKey(public.Bytes())
	ack := NewPacket(h.ConnectionID[:], h.Sequence+1, h.Channel, f)
	if _, err := c.conn.WriteTo(ack.Bytes(), addr); err != nil {
		return nil, err
	}
	s := NewSess(c.conn, addr, secret)
	s.ConnectionID = h.ConnectionID
	s.Sequence = h.Sequence
	return s, nil
}

func (c *Conn) dataHandler(sess *Sess, frame *DataFrame) error {
	data, err := frame.RawData()
	if err != nil {
		return err
	}
	if frame.Length == len(data) {
		sess.incoming <- data
		return nil
	}
	if _, ok := sess.data[frame.StreamID]; !ok {
		sess.data[frame.StreamID] = buffer.NewBuffer(frame.Length)
	}
	buf, ok := sess.data[frame.StreamID]
	if !ok {
		return nil
	}
	buf.WriteBytes(data, frame.Offset)
	if frame.Length == buf.Size() {
		delete(sess.data, frame.StreamID)
		sess.incoming <- buf.Bytes()
	}
	return nil
}

func (c *Conn) createToken(addr net.Addr) [16]byte {
	fmt.Println(addr)
	return [16]byte{}
}

func (c *Conn) verifyToken(token [16]byte) bool {
	return true
}

func (c *Conn) Close() {
	close(c.quit)
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Conn) Accept() (*Sess, error) {
	select {
	case sess := <-c.accepting:
		return sess, nil
	}
}

func Listen(addr string) (*Conn, error) {
	udpAddr, err := net.ResolveUDPAddr(protocol, addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP(protocol, udpAddr)
	if err != nil {
		return nil, err
	}
	c := &Conn{
		conn:     conn,
		sessions: sync.Map{},
		bytePool: sync.Pool{
			New: func() interface{} {
				return make([]byte, bufferSize)
			},
		},
		accepting: make(chan *Sess, queueSize),
		quit:      make(chan struct{}),
	}
	c.listen()
	return c, nil
}

func Dial(network, addr string) (*Sess, error) {
	udpAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP(udpAddr.Network(), nil, udpAddr)
	if err != nil {
		return nil, err
	}
	s := NewSess(conn, udpAddr, nil)
	s.dialer = true
	return s, acceptDial(s)
}

func acceptDial(sess *Sess) error {
	uid := uuid.NewV4().Bytes()
	sess.setConnectionID(uid)

	init := &InitFrame{
		StreamID: rand.Uint32(),
		Version:  1,
	}
	packet := NewPacket(uid, sess.NextSequence(), 1, init)
	if err := sess.send(packet.Bytes()); err != nil {
		return err
	}

	buf, err := sess.read()
	if err != nil {
		return err
	}

	header, err := DecodePacketHeader(buf)
	if err != nil {
		return err
	}
	frame := DecodeFrame(header.Type, buf[header.Size():])
	initAck := frame.(*InitAckFrame)

	private, public, err := crypto.GenerateKeys()
	if err != nil {
		return err
	}

	session := &SessionFrame{
		StreamID: rand.Uint32(),
		Token:    initAck.Token,
	}
	copy(session.Key[:], public.Bytes())
	packet = NewPacket(uid, sess.NextSequence(), 1, session)
	if err := sess.send(packet.Bytes()); err != nil {
		return err
	}

	buf, err = sess.read()
	if err != nil {
		return err
	}

	header, err = DecodePacketHeader(buf)
	if err != nil {
		return err
	}
	frame = DecodeFrame(header.Type, buf[header.Size():])
	sessAck := frame.(*SessAckFrame)

	pk := crypto.GeneratePublicKey(sessAck.Key[:])
	secret, err := crypto.ComputeSecret(private, pk)
	if err != nil {
		return err
	}
	sess.secretKey = secret
	return nil
}
