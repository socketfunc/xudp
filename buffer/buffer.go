package buffer

type Buffer struct {
	buf          []byte
	current, max int
}

func NewBuffer(size int) *Buffer {
	b := &Buffer{}
	b.buf = make([]byte, size)
	b.max = size
	return b
}

func (b *Buffer) WriteBytes(buf []byte, offset int) {
	l := offset + len(buf)
	if b.current < l {
		b.current = l
	}
	copy(b.buf[offset:], buf)
}

func (b *Buffer) Size() int {
	size := b.current
	if b.max < size {
		size = b.max
	}
	return size
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

func Iterator(buf []byte, size int, fn func(idx int, buf []byte) error) error {
	var err error
	for i := 0; i < len(buf); i += size {
		if len(buf) < i+size {
			err = fn(i, buf[i:])
		} else {
			err = fn(i, buf[i:i+size])
		}
		if err != nil {
			return err
		}
	}
	return nil
}
