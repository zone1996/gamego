package netya

import (
	"bytes"
	"fmt"

	proto "github.com/golang/protobuf/proto"
	log "github.com/zone1996/logo"
)

const MAX_PACKET_SIZE int = 1024 * 4 // PbMsg最大长度

type Codec interface {
	Encode(*PbMsg) ([]byte, bool)
	Decode(*bytes.Buffer) ([]*PbMsg, bool)
}

type DefaultCodec struct{}

func (c *DefaultCodec) Decode(in *bytes.Buffer) ([]*PbMsg, bool) {
	prelen := in.Len()
	var msgs []*PbMsg = nil
	for {
		if msg := doDecode(in); msg != nil {
			if msgs == nil {
				msgs = make([]*PbMsg, 1)
			}
			msgs = append(msgs, msg)
		} else {
			break
		}
	}
	afterLen := in.Len()
	if prelen != afterLen {
		temp := in.Bytes()
		in.Reset()
		in.Write(temp)
	}
	return msgs, msgs == nil
}

func doDecode(in *bytes.Buffer) *PbMsg {
	fmt.Println("All bytes=", in.Bytes())
	prelen := in.Len()
	log.Info("PreLen=?", prelen)
	length := readRawVarint32(in)
	if length <= 0 {
		return nil
	}
	log.Info("length=?", length)
	afterLen := in.Len()
	if afterLen < length { // not enough data for a full PbMsg
		for i := 0; i <= prelen-afterLen; i++ {
			in.UnreadByte()
		}
		return nil
	}
	pbMsgHeadLen := prelen - afterLen
	log.Info("headLen=?", pbMsgHeadLen)
	for i := 1; i <= pbMsgHeadLen; i++ {
		in.UnreadByte()
	}

	data := make([]byte, pbMsgHeadLen+length)
	in.Read(data)

	msg := &PbMsg{}
	if err := proto.Unmarshal(data, msg); err != nil {
		log.Error("PbMsg decode error:?", err.Error())
		return nil
	}
	if length > MAX_PACKET_SIZE {
		log.Info("PbMsg is too big, code=?, userId=?", msg.Code, msg.UserId)
	}
	return msg
}

// see https://github.com/netty/netty/blob/master/codec/src/main/java/io/netty/handler/codec/protobuf/ProtobufVarint32FrameDecoder.java
func readRawVarint32(in *bytes.Buffer) int {
	if in.Len() < 1 {
		return 0
	}
	unreadNum := 0
	var result int = 0
	defer func() {
		if unreadNum > 0 {
			for unreadNum != 0 {
				in.UnreadByte()
				unreadNum--
			}
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
