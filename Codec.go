package main

import (
	"encoding/binary"
	"errors"
	"gamego/netya"
	"gamego/pb"

	_ "github.com/zone1996/logo"
)

var ErrTooLargeMsg = errors.New("Too Large PbMsg")
var ErrMagicNotRight = errors.New("Magic num not Right")

// using for tcp acceptor
type Codec interface {
	Encode(*pb.PbMsg) ([]byte, bool)
	Decode(*netya.ByteBuf) ([]*pb.PbMsg, error)
}

// implements Codec
type DefaultCodec struct{}

func (c *DefaultCodec) Decode(in *netya.ByteBuf) (msgs []*pb.PbMsg, err error) {
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

func doDecode(in *netya.ByteBuf) (*pb.PbMsg, error) {
	magicBytes := in.ReadSilceN(2)
	if magicBytes == nil {
		return nil, nil
	}
	if pb.MAGIC_NUM != binary.BigEndian.Uint16(magicBytes) {
		return nil, ErrMagicNotRight // Magic Num not correct
	}
	lenBytes := in.ReadSilceN(2)
	if lenBytes == nil {
		in.UnreadBytes(2) // unread magic num bytes
		return nil, nil
	}
	length := binary.BigEndian.Uint16(lenBytes)
	if length > pb.MAX_PACKET_SIZE {
		return nil, ErrTooLargeMsg
	}
	if in.Len() < int(uint32(length)) { // decode next time
		in.UnreadBytes(4)
		return nil, nil
	}
	data := in.ReadSilceN(int(length))
	pbmsg := &pb.PbMsg{}
	err := pbmsg.ParseFrom(data)
	return pbmsg, err
}

func (c *DefaultCodec) Encode(msg *pb.PbMsg) ([]byte, bool) {
	data := msg.Bytes()
	return data, data != nil
}
