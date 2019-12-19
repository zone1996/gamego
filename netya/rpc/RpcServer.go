package rpc

import (
	"gamego/netya"

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

type RpcCall int

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
	cs := "0.0"
	serviceName := strings.Split(cs, ".")[0]
	methodName := strings.Split(cs, ".")[1]

	service, ok := rpch.services.Load(serviceName)
	if !ok {
		// handle service not found
		return
	}
	method := reflect.ValueOf(service).MethodByName(methodName)            // 获取方法
	mt := method.Type()                                                    // 方法type
	in0Type := mt.In(0)                                                    // 参数0类型
	in1Type := mt.In(1)                                                    // 参数1类型
	in0Val := reflect.New(in0Type)                                         // 实例化参数0 Req
	in1Val := reflect.New(in1Type)                                         // 实例化参数1 Resp
	proto.Unmarshal([]byte(""), in0Val.Elem().Interface().(proto.Message)) // 解码Req
	method.Call([]reflect.Value{in0Val, in1Val})                           // 调用
	resp, err := proto.Marshal(in1Val.Elem().Interface().(proto.Message))  // 编码Resp
	// call.resp = resp
	// session.Write(call)
	session.Write(resp) // 回传
}

func decode(in *netya.ByteBuf) ([]*RpcCall, error) {
	return nil, nil
}
