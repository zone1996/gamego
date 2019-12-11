package netya

import (
	"encoding/binary"

	log "github.com/zone1996/logo"

	"github.com/golang/protobuf/proto"
)

const MAX_PACKET_SIZE = ^uint16(0) >> 1 // 0x7FFF
const MAGIC_NUM = uint16(0x1234)

type PbMsg struct {
	length uint16
	msg    *Msg
}

func NewPbMsg(code int32) *PbMsg {
	return &PbMsg{
		msg: &Msg{
			Code: code,
		},
	}
}

func (this *PbMsg) SetCode(code int32) {
	this.msg.Code = code
}

func (this *PbMsg) GetCode() int32 {
	return this.msg.Code
}

func (this *PbMsg) SetUserId(userId int32) {
	this.msg.UserId = userId
}

func (this *PbMsg) GetUserId() int32 {
	return this.msg.GetUserId()
}

func (this *PbMsg) SetValue(i, value int32) {
	switch i {
	case 0:
		this.msg.Value0 = value
	case 1:
		this.msg.Value1 = value
	case 2:
		this.msg.Value2 = value
	}
}

func (this *PbMsg) GetValue(i int32) int32 {
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

func (this *PbMsg) GetPayload() []byte {
	return this.msg.GetPayload()
}

func (this *PbMsg) SetPayload(payload []byte) {
	this.msg.Payload = payload
}

func (this *PbMsg) Bytes() []byte {
	md, err := proto.Marshal(this.msg)
	if err != nil {
		log.Info("ProtoBuf Marshal err:?", err)
		return nil
	}
	if this.msg.XXX_Size() > int(MAX_PACKET_SIZE) {
		log.Info("PbMsg too large, code=?", this.GetCode())
		return nil
	}
	this.length = uint16(this.msg.XXX_Size())
	b := make([]byte, this.length+4)
	binary.BigEndian.PutUint16(b, MAGIC_NUM)
	binary.BigEndian.PutUint16(b[2:], this.length)
	copy(b[4:], md)
	return b
}

func (this *PbMsg) ParseFrom(p []byte) error {
	m := &Msg{}
	err := proto.Unmarshal(p, m)
	if err != nil {
		return err
	}
	this.msg = m
	return nil
}

func (this *PbMsg) Clone() *PbMsg {
	clone := &PbMsg{
		msg: &Msg{
			Code:   this.GetCode(),
			UserId: this.GetUserId(),
			Value0: this.GetValue(0),
			Value1: this.GetValue(1),
			Value2: this.GetValue(2),
		},
	}
	payload := make([]byte, len(this.GetPayload()))
	clone.SetPayload(payload)
	return clone
}
