package netya

import (
	"encoding/binary"
	"errors"
	"fmt"

	_ "github.com/zone1996/logo"
)

var ErrTooLargeMsg = errors.New("Too Large PbMsg")
var ErrMagicNotRight = errors.New("Magic num not Right")

type Codec interface {
	Encode(*PbMsg) ([]byte, bool)
	Decode(*ByteBuf) ([]*PbMsg, error)
}

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
	fmt.Println("in bytes:", in.Bytes())
	magicBytes := in.ReadSilceN(2)
	if magicBytes == nil {
		return nil, nil
	}
	if MAGIC_NUM != binary.BigEndian.Uint16(magicBytes) {
		fmt.Println("MagicByte:", magicBytes)
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
