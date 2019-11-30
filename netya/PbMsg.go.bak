package netya

import (
	_ "github.com/golang/protobuf/proto"
)

const HEADER_SIZE int = 24
const MAGIC_NUM int16 = 0x1234

type PbMsg struct {
	// header：24字节
	magicNum int16 // 魔数
	code     int16 // 协议号
	length   int16 // 消息体长度
	checkSum int32 // 校验和
	userId   int32 // 用户ID
	Value0   int32 // 可选值
	Value1   int32 // 可选值
	Value2   int32 // 可选值

	payload []byte      // 消息体
	pb      interface{} // 消息体反序列化得到，程序内使用，不参与数据传输
}

func (PbMsg) NewPbMsg(code int16) (pbmsg *PbMsg) {
	pbmsg = &PbMsg{
		magicNum: MAGIC_NUM,
		code:     code,
	}
	return
}

func (this *PbMsg) SetCode(code int16) {
	this.code = code
}

func (this *PbMsg) GetCode() int16 {
	return this.code
}

func (this *PbMsg) SetUserId(userId int32) {
	this.userId = userId
}

func (this *PbMsg) GetUserId() int32 {
	return this.userId
}

func (this *PbMsg) GetMagicNum() int16 {
	return this.magicNum
}

func (this *PbMsg) GetCheckSum() int32 {
	return this.checkSum
}

func (this *PbMsg) SetValue(i, value int32) {
	switch i {
	case 0:
		this.Value0 = value
	case 1:
		this.Value1 = value
	case 2:
		this.Value2 = value
	}
}

func (this *PbMsg) GetValue(i int32) int32 {
	switch i {
	case 0:
		return this.Value0
	case 1:
		return this.Value1
	case 2:
		return this.Value2
	}
	return 0
}

func (this *PbMsg) GetPayload() []byte {
	return this.payload
}

func (this *PbMsg) SetPayload(payload []byte) {
	this.payload = make([]byte, len(payload))
	copy(this.payload, payload)
}

func (this *PbMsg) GetPB() interface{} {
	return this.pb
}

func (this *PbMsg) Clone() *PbMsg {
	clone := &PbMsg{
		magicNum: MAGIC_NUM,
		code:     this.code,
		length:   this.length,
		checkSum: this.checkSum,
		userId:   this.userId,
		Value0:   this.Value0,
		Value1:   this.Value1,
		Value2:   this.Value2,

		payload: make([]byte, len(this.payload)),
		pb:      this.pb,
	}
	copy(clone.payload, this.payload)
	return clone
}
