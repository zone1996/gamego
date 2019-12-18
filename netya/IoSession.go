package netya

type IoSession interface {
	Write(b []byte) (n int, err error) // 同步写
	WriteAsync(b []byte)               // 异步写

	Closed() bool
	Close()

	SetAttribute(k string, v interface{})
	GetAttribute(k string) interface{}
}
