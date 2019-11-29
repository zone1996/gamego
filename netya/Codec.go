package netya

type Codec interface {
	Encode(PbMsg) []byte
	Decode([]byte) ([]PbMsg, bool)
}

type DefaultCodec struct{}

// TODO 将PbMsg改为protobuf实现
func (c *DefaultCodec) Decode(data []byte) ([]PbMsg, bool) {
	if len(data) < HEADER_SIZE {
		return nil, false
	}
	
}


func (c *DefaultCodec) Encode(msg PbMsg) []byte {
	data := make(byte[], HEADER_SIZE + len(msg.GetPayload()))
	
	return data
}