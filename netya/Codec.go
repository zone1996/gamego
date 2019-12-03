package netya

import (
	"errors"

	proto "github.com/golang/protobuf/proto"
	log "github.com/zone1996/logo"
)

const MAX_PACKET_SIZE int = 1024 * 4 // PbMsg最大长度
var ErrTooLargeMsg = errors.New("Too Large PbMsg")

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
	prelen := in.Len()
	length := readRawVarint32(in)
	if length <= 0 {
		return nil, nil
	}
	afterLen := in.Len()
	if afterLen < length { // not enough data for a full PbMsg
		in.UnreadBytes(prelen - afterLen)
		return nil, nil
	}
	pbMsgHeadLen := prelen - afterLen
	in.UnreadBytes(pbMsgHeadLen)
	data := in.ReadSilceN(pbMsgHeadLen + length)
	msg := &PbMsg{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	if length > MAX_PACKET_SIZE {
		log.Info("PbMsg is too big, code=?, userId=?", msg.Code, msg.UserId)
		if length > MAX_PACKET_SIZE*2 {
			return nil, ErrTooLargeMsg
		}
	}
	return msg, nil
}

// see https://github.com/netty/netty/blob/master/codec/src/main/java/io/netty/handler/codec/protobuf/ProtobufVarint32FrameDecoder.java
func readRawVarint32(in *ByteBuf) int {
	if in.Len() < 1 {
		return 0
	}
	_, _ = in.ReadByte()
	unreadNum := 1
	var result int = 0
	defer func() {
		if unreadNum > 1 {
			in.UnreadBytes(unreadNum)
		}
	}()

	tmp, _ := in.ReadByte() // 1
	if tmp >= 0 {           // if tmp == 0 , discard this byte
		return int(tmp)
	} else {
		tmp &= 127
		result |= int(tmp)
		if in.Len() < 1 { // read next time
			unreadNum++
			return 0
		}
	}

	tmp, _ = in.ReadByte() // 2
	if tmp >= 0 {
		result |= int(tmp) << 7
		return result
	} else {
		tmp &= 127
		result |= int(tmp) << 7
		if in.Len() < 1 {
			unreadNum++
			return 0
		}
	}

	tmp, _ = in.ReadByte() // 3
	if tmp >= 0 {
		result |= int(tmp) << 14
		return result
	} else {
		tmp &= 127
		result |= int(tmp) << 14
		if in.Len() < 1 {
			unreadNum++
			return 0
		}
	}

	tmp, _ = in.ReadByte() // 4
	if tmp >= 0 {
		result |= int(tmp) << 21
		return result
	} else {
		tmp &= 127
		result |= int(tmp) << 21
		if in.Len() < 1 {
			unreadNum++
			return 0
		}
	}

	tmp, _ = in.ReadByte() // 5
	if tmp < 0 {
		unreadNum = 0
		panic("malformed varint.")
	}
	tmp &= 127
	result |= int(tmp) << 28
	return result
}

func (c *DefaultCodec) Encode(msg *PbMsg) ([]byte, bool) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Error("PbMsg encode error:?", err)
	}
	return data, err == nil
}
