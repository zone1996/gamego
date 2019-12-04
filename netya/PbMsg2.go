package netya

import (
	"encoding/binary"

	log "github.com/zone1996/logo"

	"github.com/golang/protobuf/proto"
)

const MAX_LENGTH = ^uint16(0) >> 1
const MAGIC_NUM = uint16(0x1234)

type PbMsg2 struct {
	length uint16
	msg    *PbMsg
}

func NewPbMsg2(code int32) *PbMsg2 {
	return &PbMsg2{
		msg: &PbMsg{
			code: code,
		},
	}
}

func (this *PbMsg2) SetCode(code int32) {
	this.msg.Code = code
}

func (this *PbMsg2) GetCode() int32 {
	return this.msg.Code
}

func (this *PbMsg2) SetUserId(userId int32) {
	this.msg.UserId = userId
}

func (this *PbMsg2) GetUserId() int32 {
	return this.msg.GetUserId()
}

func (this *PbMsg2) SetValue(i, value int32) {
	switch i {
	case 0:
		this.msg.Value0 = value
	case 1:
		this.msg.Value1 = value
	case 2:
		this.msg.Value2 = value
	}
}

func (this *PbMsg2) GetValue(i int32) int32 {
	switch i {
	case 0:
		return this.msg.GetValue0()
	case 1:
		return this.msg.GetValue1()
	case 2:
		return this.msg.GetValue2()
	}
	return 0
}

func (this *PbMsg2) GetPayload() []byte {
	return this.msg.GetPayload()
}

func (this *PbMsg2) SetPayload(payload []byte) {
	this.msg.Payload = payload
}

func (this *PbMsg2) Bytes() []byte {
	md, err := proto.Marshal(this.msg)
	if err != nil {
		log.Info("ProtoBuf Marshal err:?", err)
		return nil
	}
	this.length = 4 + this.msg.XXX_Size()
	if this.length > MAX_LENGTH {
		log.Info("PbMsg too large, code=?", this.GetCode())
		return nil
	}
	b := make([]byte, this.length)
	binary.LittleEndian.PutUint16(b, MAGIC_NUM)
	binary.LittleEndian.PutUint16(b[2:], this.length)
	copy(b[4:], md)
	return b
}

func (this *PbMsg2) Clone() *PbMsg2 {
	clone := &PbMsg2{
		msg: &PbMsg{
			code:   this.GetCode(),
			userId: this.GetUserId(),
			Value0: this.GetValue(0),
			Value1: this.GetValue(1),
			Value2: this.GetValue(2),
		},
	}
	payload := make([]byte, len(this.GetPayload())),
		clone.SetPayload(payload)
	return clone
}
