package crypto

import (
	"crypto"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"

	"github.com/aead/ecdh"
)

var p256 = ecdh.Generic(elliptic.P256())

type PrivateKey struct {
	key crypto.PrivateKey
}

func (p *PrivateKey) Bytes() []byte {
	switch t := p.key.(type) {
	case []byte:
		return t
	case *[]byte:
		return *t
	}
	return nil
}

type PublicKey struct {
	key crypto.PublicKey
}

func (p *PublicKey) Bytes() []byte {
	switch t := p.key.(type) {
	case ecdh.Point:
		x := t.X.Bytes()
		y := t.Y.Bytes()
		buf := make([]byte, 0, len(x)+len(y))
		buf = append(buf, x...)
		buf = append(buf, y...)
		return buf
	case *ecdh.Point:
		x := t.X.Bytes()
		y := t.Y.Bytes()
		buf := make([]byte, 0, len(x)+len(y))
		buf = append(buf, x...)
		buf = append(buf, y...)
		return buf
	}
	return nil
}

func GenerateKeys() (*PrivateKey, *PublicKey, error) {
	private, public, err := p256.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return &PrivateKey{key: private}, &PublicKey{key: public}, nil
}

func GeneratePublicKey(buf []byte) *PublicKey {
	x := &big.Int{}
	x = x.SetBytes(buf[0:32])
	y := &big.Int{}
	y = y.SetBytes(buf[32:64])
	point := ecdh.Point{
		X: x,
		Y: y,
	}
	pk := crypto.PublicKey(point)
	return &PublicKey{key: pk}
}

func ComputeSecret(private *PrivateKey, public *PublicKey) ([]byte, error) {
	if err := p256.Check(public.key); err != nil {
		return nil, err
	}
	return p256.ComputeSecret(private.key, public.key), nil
}
