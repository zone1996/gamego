package rpc

import (
	"encoding/binary"
	"gamego/netya"

	log "github.com/zone1996/logo"

	"github.com/golang/protobuf/proto"

	"errors"
	"reflect"
	"strings"
	"sync"
)

type RpcServer struct {
	acceptor netya.Acceptor
	services *sync.Map // string:interface{}, struct实例
}

func NewRpcServer(config *netya.AcceptorConfig) *RpcServer {
	services := &sync.Map{}
	s := &RpcServer{
		acceptor: netya.NewTCPAcceptor(config, &rpcHandler{services}),
		services: services,
	}
	return s
}

var ErrServiceDuplicated = errors.New("Rpc service duplicated")
var ErrServiceNotStruct = errors.New("Rpc service not struct")
var ErrServiceNotFound = errors.New("Rpc service not found")

func (server *RpcServer) Register(service interface{}) (err error) {
	v := reflect.ValueOf(service)
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	v.NumField() // mustBe struct
	serviceName := reflect.Indirect(v).Type().Name()
	if _, ok := server.services.Load(serviceName); ok {
		return ErrServiceDuplicated // 重复注册
	}
	server.services.LoadOrStore(serviceName, v)
	return nil
}

func (server *RpcServer) Unregister(serviceName string) {
	server.services.Delete(serviceName)
}

func (server *RpcServer) Start() {
	go server.acceptor.Accept()
}

func (server *RpcServer) Stop() {
	server.acceptor.Shutdown()
}

type rpcHandler struct {
	services *sync.Map
}

func (rpch *rpcHandler) OnConnected(session netya.IoSession) {}

func (rpch *rpcHandler) OnDisconnected(session netya.IoSession) {}

func (rpch *rpcHandler) OnMessage(session netya.IoSession, message []byte) {
	inByteBuf := session.(*netya.TCPSession).InBoundBuffer
	_, err := inByteBuf.Write(message)
	if err != nil {
		session.Close()
		return
	}

	if calls, err := decode(inByteBuf); calls != nil && err == nil {
		for _, call := range calls {
			rpch.handleService(session, call)
		}
	} else {
		session.Close()
		return
	}
}

func (rpch *rpcHandler) handleService(session netya.IoSession, call *RpcCall) {
	var err error
	var in1Val reflect.Value
	for {
		sm := strings.Split(call.ServiceName, ".")
		if len(sm) != 2 {
			err = ErrServiceNotFound
			break
		}
		service, ok := rpch.services.Load(sm[0])
		if !ok { // 服务未注册
			err = ErrServiceNotFound
			break
		}
		method := reflect.ValueOf(service).MethodByName(sm[1]) // 获取方法
		if method.Elem().Interface() == nil {                  // 方法不存在
			err = ErrServiceNotFound
			break
		}
		mt := method.Type()                                                          // 方法type
		in0Type := mt.In(0)                                                          // 参数0类型
		in1Type := mt.In(1)                                                          // 参数1类型
		in0Val := reflect.New(in0Type)                                               // 实例化参数0 Req
		in1Val = reflect.New(in1Type)                                                // 实例化参数1 Resp
		err = proto.Unmarshal([]byte(""), in0Val.Elem().Interface().(proto.Message)) // 解码Req
		if err != nil {
			break
		}
		err = method.Call([]reflect.Value{in0Val, in1Val})[0].Elem().Interface().(error) // 调用
		break
	}
	if err != nil {
		log.Info("RpcCall service=?, err:?", call.ServiceName, err.Error())
	}
	if call.GetReply() {
		call.Req = nil
		if err != nil {
			call.Err = err.Error()
		} else {
			if resp, err := proto.Marshal(in1Val.Elem().Interface().(proto.Message)); err == nil {
				call.Resp = resp
			} else {
				call.Err = err.Error()
				log.Info("RpcCall service=?, err:?", call.ServiceName, err.Error())
			}
		}
		back, _ := proto.Marshal(call)
		session.Write(back) // 回传
	}
}

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

const MAGIC_NUM = uint16(0x1234)
const MAX_PACKET_SIZE = ^uint16(0) >> 1 // 0x7FFF=32767
var ErrTooLargeMsg = errors.New("Too Large PbMsg")
var ErrMagicNotRight = errors.New("Magic num not Right")

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
