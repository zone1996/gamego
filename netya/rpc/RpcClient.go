package rpc

import (
	"errors"
	"gamego/netya"
	"sync"
	"time"

	log "github.com/zone1996/logo"

	"github.com/golang/protobuf/proto"
)

var ErrRpcCallTimedOut = errors.New("RPC Call timedout")

type RpcClient struct {
	seqGen         int64
	connector      netya.Connector
	mu             sync.Mutex
	calls          *sync.Map // seq:*rpcCallRespHolder
	callWaitChanns *sync.Map //seq:chan struct{} using for block-method call
}

type rpcCallRespHolder struct {
	seq       int64
	resp      []byte
	err       error
	respBytes chan []byte
}

func NewRpcClient(addr string) *RpcClient {
	calls, callWaitChanns := &sync.Map{}, &sync.Map{}
	h := &rpcClientHandler{
		calls:          calls,
		callWaitChanns: callWaitChanns,
	}
	c := &RpcClient{
		connector:      netya.NewTCPConnector(addr, h),
		calls:          calls,
		callWaitChanns: callWaitChanns,
	}
	if ok := c.connector.Connect(); !ok {
		return nil
	}
	return c
}

func (client *RpcClient) Call(service string, req, resp proto.Message) error {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	call := &RpcCall{}
	call.Seq = client.genSeq()
	call.ServiceName = service
	call.Req = reqBytes
	call.Reply = true

	callBytes, err := encode(call) // 编码
	if err != nil {
		return err
	}

	holder := &rpcCallRespHolder{
		seq: call.Seq,
	}
	client.calls.Store(call.Seq, holder)
	waitChan := make(chan struct{})
	client.callWaitChanns.Store(call.Seq, waitChan)

	if _, err := client.write(callBytes); err != nil { // 发送
		client.calls.Delete(call.Seq)
		client.callWaitChanns.Delete(call.Seq)
		return err
	}
	<-waitChan // 等待结果
	if holder.err != nil {
		return holder.err
	}
	return proto.Unmarshal(holder.resp, resp) // 解码
}

func (client *RpcClient) CallAsync(service string, req proto.Message) (<-chan []byte, error) {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	call := &RpcCall{}
	call.Seq = client.genSeq()
	call.ServiceName = service
	call.Req = reqBytes
	call.Reply = true

	callBytes, err := encode(call) // 编码
	if err != nil {
		return nil, err
	}

	ch := make(chan []byte)
	holder := &rpcCallRespHolder{
		seq:       call.Seq,
		respBytes: ch,
	}
	client.calls.Store(call.Seq, holder)
	if _, err := client.write(callBytes); err != nil { // 发送
		client.calls.Delete(call.Seq)
		return nil, err
	}
	return ch, nil
}

func (client *RpcClient) CallTimedOut(service string,
	req, resp proto.Message, timedOutNanos time.Duration, failRetryCount int) error {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	call := &RpcCall{}
	call.Seq = client.genSeq()
	call.ServiceName = service
	call.Req = reqBytes
	call.Reply = true

	callBytes, err := encode(call) // 编码
	if err != nil {
		return err
	}

	holder := &rpcCallRespHolder{
		seq: call.Seq,
	}
	client.calls.Store(call.Seq, holder)
	waitChan := make(chan struct{})
	client.callWaitChanns.Store(call.Seq, waitChan)

	if _, err := client.write(callBytes); err != nil { // 发送
		client.calls.Delete(call.Seq)
		client.callWaitChanns.Delete(call.Seq)
		return err
	}

	timedOut := time.After(timedOutNanos)
	select {
	case <-waitChan:
		if holder.err != nil {
			return holder.err
		}
		return proto.Unmarshal(holder.resp, resp)
	case <-timedOut:
		client.calls.Delete(call.Seq)
		client.callWaitChanns.Delete(call.Seq)
		log.Info("RpcCall timed out, service:?", service)
		return ErrRpcCallTimedOut
	}
}

func (client *RpcClient) CallNoReply(service string, req proto.Message) error {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	call := &RpcCall{}
	// call.Seq = client.genSeq()
	call.ServiceName = service
	call.Req = reqBytes
	call.Reply = false

	callBytes, err := encode(call) // 编码
	if err != nil {
		return err
	}
	_, err = client.write(callBytes) // 发送
	return err
}

func (client *RpcClient) genSeq() int64 {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.seqGen++
	return client.seqGen
}

func (client *RpcClient) write(b []byte) (n int, err error) {
	return client.connector.Write(b)
}

//////////////////////////////////////////////

// rpcClientHandler implement netya.IoHandler
type rpcClientHandler struct {
	netya.IoHandlerAdapter
	calls          *sync.Map // seq:*rpcCallRespHolder
	callWaitChanns *sync.Map
}

func (rpch *rpcClientHandler) OnMessage(session netya.IoSession, message []byte) {
	inByteBuf := session.(*netya.TCPSession).InBoundBuffer
	_, err := inByteBuf.Write(message)
	if err != nil {
		session.Close()
		return
	}

	if calls, err := decode(inByteBuf); calls != nil && err == nil {
		for _, call := range calls {
			rpch.handleResult(call)
		}
	} else if err != nil {
		session.Close()
		return
	}
}

func (rpch *rpcClientHandler) handleResult(call *RpcCall) {
	holder, ok := rpch.calls.Load(call.Seq)
	if ok {
		respHolder := holder.(*rpcCallRespHolder)
		if respHolder.respBytes != nil { // 异步调用
			respHolder.respBytes <- call.GetResp()
			close(respHolder.respBytes)
		} else {
			respHolder.resp = call.GetResp() // 同步调用
		}
		if []byte(call.Err) != nil {
			respHolder.err = errors.New(call.Err)
		}
		rpch.calls.Delete(call.Seq)
		if ch, ok := rpch.callWaitChanns.Load(call.Seq); ok {
			ch.(chan struct{}) <- struct{}{}
			rpch.callWaitChanns.Delete(call.Seq)
		}
	} else {
		log.Info("Service not found:?", call.ServiceName)
	}
}

func (rpch *rpcClientHandler) OnDisconnected(session netya.IoSession) {
	// TODO 为rpch.calls中的holder设置连接关闭错误
	// TODO 关闭rpch.callWaitChanns中的chan
}
