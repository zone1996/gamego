package rpc

import (
	"encoding/binary"
	"errors"
	"gamego/netya"

	"github.com/golang/protobuf/proto"
)

const MAGIC_NUM = uint16(0x1234)
const MAX_PACKET_SIZE = ^uint16(0) >> 1 // 0x7FFF=32767
var ErrTooLargeMsg = errors.New("Too Large PbMsg")
var ErrMagicNotRight = errors.New("Magic num not Right")

func decode(in *netya.ByteBuf) (calls []*RpcCall, err error) {
	for {
		if call, e := doDecode(in); call != nil {
			calls = append(calls, call)
		} else {
			err = e
			break
		}
	}
	return
}

func doDecode(in *netya.ByteBuf) (*RpcCall, error) {
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
	call := &RpcCall{}
	err := proto.Unmarshal(data, call)
	return call, err
}
