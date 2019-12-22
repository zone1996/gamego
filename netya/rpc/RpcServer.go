package rpc

import (
	"gamego/netya"
	"time"

	log "github.com/zone1996/logo"

	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
)

var ErrServiceDuplicated = errors.New("RpcServer: service duplicated")
var ErrServiceNotStruct = errors.New("RpcServer: service not struct")
var ErrServiceNotFound = errors.New("RpcServer: service not found")
var ErrServiceNoExported = errors.New("RpcServer: No Exported Service")

type methodInfo struct {
	method   reflect.Value // 方法
	reqType  reflect.Type  // 请求参数类型
	respType reflect.Type  // 返回参数类型
}

type rpcservice struct {
	serviceName string                 // struct name
	m           map[string]*methodInfo // methodName:methodInfo
}

type RpcServer struct {
	acceptor netya.Acceptor
	services *sync.Map // string:*rpcservice
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
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	v := reflect.ValueOf(service)
	t := reflect.TypeOf(service)
	serviceName := reflect.Indirect(v).Type().Name() // struct name
	if _, ok := server.services.Load(serviceName); ok {
		return ErrServiceDuplicated // 重复注册
	}

	methodNum := t.NumMethod()
	log.Info("serviceName ?, methodNum:?", serviceName, methodNum)
	if methodNum < 1 {
		log.Info("No exported method:?", serviceName)
		return ErrServiceNoExported
	}

	s := &rpcservice{
		serviceName: serviceName,
		m:           make(map[string]*methodInfo),
	}
	for i := 0; i < methodNum; i++ {
		m := v.Method(i) // 方法m:reflect.Value
		mt := m.Type()
		mi := &methodInfo{
			method:   m,
			reqType:  mt.In(0),
			respType: mt.In(1),
		}
		s.m[t.Method(i).Name] = mi
	}
	server.services.Store(serviceName, s)
	for mn, _ := range s.m {
		log.Info("method name:?", mn)
	}
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
			b := time.Now().UnixNano()
			rpch.handleService(session, call)
			e := time.Now().UnixNano()
			log.Info("exec RpcService [?], cost ?ns", call.ServiceName, (e - b))
		}
	} else if err != nil {
		session.Close()
		return
	}
}

func (rpch *rpcServerHandler) handleService(session netya.IoSession, call *RpcCall) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("handleService recover() :?", r)
		}
	}()
	var err error
	var respVal reflect.Value
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
		mi, ok := service.(*rpcservice).m[sm[1]]
		if !ok { // 服务未注册
			err = ErrServiceNotFound
			break
		}
		reqVal := reflect.New(mi.reqType)                                   // 实例化参数0 Req
		respVal = reflect.New(mi.respType.Elem())                           // 实例化参数1 Resp
		err = proto.Unmarshal(call.Req, reqVal.Interface().(proto.Message)) // 解码Req
		if err != nil {
			break
		}

		if vs := mi.method.Call([]reflect.Value{reqVal.Elem(), respVal}); vs != nil { // 调用
			if e := vs[0].Interface(); e != nil {
				err = e.(error)
			}
		}
		break
	}

	if err != nil {
		log.Info("RpcCall service=?, err:?", call.ServiceName, err)
	}
	if call.GetReply() {
		call.Req = nil
		if err != nil {
			call.Err = err.Error()
		} else {
			if resp, err := proto.Marshal(respVal.Interface().(proto.Message)); err == nil {
				call.Resp = resp
			} else {
				call.Err = err.Error()
				log.Info("RpcCall service=?, err:?", call.ServiceName, err)
			}
		}
		if back, err := encode(call); err == nil {
			session.Write(back) // 回传
		} else {
			log.Info("Err:?", err)
		}
	}
}
