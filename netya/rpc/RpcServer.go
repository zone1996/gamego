package rpc

import (
	"gamego/netya"

	log "github.com/zone1996/logo"

	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
)

var ErrServiceDuplicated = errors.New("Rpc service duplicated")
var ErrServiceNotStruct = errors.New("Rpc service not struct")
var ErrServiceNotFound = errors.New("Rpc service not found")

type RpcServer struct {
	acceptor netya.Acceptor
	services *sync.Map // string:interface{}, struct实例
}

func NewRpcServer(config *netya.AcceptorConfig) *RpcServer {
	services := &sync.Map{}
	s := &RpcServer{
		acceptor: netya.NewTCPAcceptor(config, &rpcServerHandler{services: services}),
		services: services,
	}
	return s
}

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

///////////////////////////////////////////
//rpcServerHandler implement netya.IoHandle
type rpcServerHandler struct {
	netya.IoHandlerAdapter
	services *sync.Map
}

func (rpch *rpcServerHandler) OnMessage(session netya.IoSession, message []byte) {
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
	} else if err != nil {
		session.Close()
		return
	}
}

func (rpch *rpcServerHandler) handleService(session netya.IoSession, call *RpcCall) {
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
