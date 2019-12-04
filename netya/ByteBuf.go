package netya

import (
	"errors"
	"io"
)

type ByteBuf struct {
	buf    []byte
	r      int
	w      int
	maxCap int
}

var ErrByteBufCapNotEnough = errors.New("ByteBuf left capacity not enough")
var ErrByteBufEmpty = errors.New("ByteBuf is Empty")
var ErrByteBufUnread = errors.New("Can't unread this much bytes")

func NewByteBuf(initCap int, maxCap int) *ByteBuf {
	if initCap <= 0 {
		initCap = 64
	}
	if maxCap > 32*1024 {
		maxCap = 32 * 1024
	}
	if initCap > maxCap {
		initCap = maxCap
	}

	bb := &ByteBuf{
		buf:    make([]byte, initCap),
		r:      0,
		w:      0,
		maxCap: maxCap,
	}
	return bb
}

func (bb *ByteBuf) Empty() bool {
	return bb.r == bb.w
}

// not safe read
func (bb *ByteBuf) ReadSilceN(n int) []byte {
	if n <= 0 || bb.Len() < n {
		if bb.Empty() {
			bb.Reset()
		}
		return nil
	}
	p := bb.buf[bb.r : bb.r+n]
	bb.r += n
	return p
}

func (bb *ByteBuf) Read(p []byte) (int, error) {
	lp := len(p)
	if bb.Empty() {
		bb.Reset()
		if lp == 0 {
			return 0, io.EOF
		}
		return 0, nil
	}
	n := lp
	if bb.Len() < lp {
		n = bb.Len()
	}
	n = copy(p, bb.buf[bb.r:bb.r+n])
	bb.r += n
	return n, nil
}

func (bb *ByteBuf) ReadByte() (byte, error) {
	if bb.Empty() {
		return 0, ErrByteBufEmpty
	}
	b := bb.buf[bb.r]
	bb.r++
	return b, nil
}

func (bb *ByteBuf) UnreadBytes(n int) error {
	if bb.r < n {
		return ErrByteBufUnread
	}
	bb.r -= n
	return nil
}

func (bb *ByteBuf) Bytes() []byte {
	return bb.buf[bb.r:bb.w]
}

func (bb *ByteBuf) BytesN(n int) []byte {
	if n <= 0 || bb.Len() < n {
		return nil
	}
	return bb.buf[bb.r : bb.r+n]
}

func (bb *ByteBuf) Reset() {
	bb.buf = bb.buf[:0]
	bb.r = 0
	bb.w = 0
}

func (bb *ByteBuf) WriteByte(p byte) error {
	if !bb.tryGrow(1) {
		return ErrByteBufCapNotEnough
	}
	bb.buf[bb.w] = p
	bb.w++
	return nil
}

func (bb *ByteBuf) Write(p []byte) (int, error) {
	if !bb.tryGrow(len(p)) {
		return 0, ErrByteBufCapNotEnough
	}
	n := copy(bb.buf[bb.w:], p)
	bb.w += n
	return n, nil
}

func (bb *ByteBuf) Cap() int {
	return cap(bb.buf)
}

func (bb *ByteBuf) Len() int {
	return bb.w - bb.r
}

// grow when not enough space to write
func (bb *ByteBuf) tryGrow(n int) bool {
	if bb.w+n <= bb.Cap() {
		return true
	}
	if bb.r >= n {
		bb.shiftToHead()
		return true
	}

	if bb.Cap()+n > bb.maxCap {
		return false
	}
	// reallocate a []byte
	nc := bb.Cap()*2 + n
	if nc > bb.maxCap {
		nc = bb.maxCap
	}
	buf := make([]byte, nc)
	copy(buf, bb.buf[bb.r:bb.w])
	bb.buf = buf
	bb.w = bb.Len()
	bb.r = 0
	return true
}

func (bb *ByteBuf) shiftToHead() {
	copy(bb.buf, bb.buf[bb.r:bb.w])
	bb.w = bb.Len()
	bb.r = 0
}
