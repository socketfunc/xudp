package xudp

import (
	"testing"
)

func TestNewSess(t *testing.T) {
	sess := NewSess(nil, nil, nil)
	defer sess.Close()

	buf := make([]byte, 1024)
	sess.Send(buf)
}
