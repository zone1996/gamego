package netya

import (
	"encoding/binary"
	"errors"

	_ "github.com/zone1996/logo"
)

var ErrTooLargeMsg = errors.New("Too Large PbMsg")
var ErrMagicNotRight = errors.New("Magic num not Right")

// using for tcp acceptor
type Codec interface {
	Encode(*PbMsg) ([]byte, bool)
	Decode(*ByteBuf) ([]*PbMsg, error)
}

// implements Codec
type DefaultCodec struct{}

func (c *DefaultCodec) Decode(in *ByteBuf) (msgs []*PbMsg, err error) {
	for {
		if msg, e := doDecode(in); msg != nil {
			msgs = append(msgs, msg)
		} else {
			err = e
			break
		}
	}
	return
}

func doDecode(in *ByteBuf) (*PbMsg, error) {
	magicBytes := in.ReadSilceN(2)
	if magicBytes == nil {
		return nil, nil
	}
	if MAGIC_NUM != binary.BigEndian.Uint16(magicBytes) {
		return nil, ErrMagicNotRight // Magic Num not correct
	}
	lenBytes := in.ReadSilceN(2)
	if lenBytes == nil {
		in.UnreadBytes(2) // unread magic num bytes
		return nil, nil
	}
	length := binary.BigEndian.Uint16(lenBytes)
	if length > MAX_PACKET_SIZE {
		return nil, ErrTooLargeMsg
	}
	if in.Len() < int(uint32(length)) { // decode next time
		in.UnreadBytes(4)
		return nil, nil
	}
	data := in.ReadSilceN(int(length))
	pbmsg := &PbMsg{}
	err := pbmsg.ParseFrom(data)
	return pbmsg, err
}

func (c *DefaultCodec) Encode(msg *PbMsg) ([]byte, bool) {
	data := msg.Bytes()
	return data, data != nil
}

type UdpCodec interface {
	Encode(*PbMsg) ([]byte, bool)
	Decode([]byte) (*PbMsg, error)
}

type DefaultUdpCodec struct{}

func (duc *DefaultUdpCodec) Decode(data []byte) (*PbMsg, error) {
	if MAGIC_NUM != binary.BigEndian.Uint16(data[:2]) {
		return nil, ErrMagicNotRight // Magic Num not correct
	}
	pbmsg := &PbMsg{}
	err := pbmsg.ParseFrom(data[4:])
	return pbmsg, err
}
func (duc *DefaultUdpCodec) Encode(msg *PbMsg) ([]byte, bool) {
	data := msg.Bytes()
	return data, data != nil
}
