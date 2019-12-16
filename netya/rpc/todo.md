添加跨语言的RPC调用。

hello.proto
package rpcpb
message HelloRequest {
	string name = 1;
}

message HelloResponse {
	string msg = 1;
}

type HelloService struct{}

func (hs *HelloService) Hello(req HelloRequest, resp *HelloResponse) {
	resp.Msg = "Hello," + req.Name;
}

Server端：

RpcServer.Register(&HelloService{})
...register more...
RpcServer.Start()


========================================================================
Client端：
RpcClient.Dial(tcp, serverAddr)

req := HelloRequest{}
req.Name = "client"
resp := &HelloResponse{}

RpcClient.Call("HelloService.Hello", req, resp)
RpcClient.CallAsync("HelloService.Hello", req) <-chan interface{}
RpcClient.CallTimedOut("HelloService.Hello", req, resp, timedOutNanos) error


==============
rpc.proto
package rpcpb
message RpcCall {
	string serviceName = 1;
	bytes req = 2;
	bytes resp = 3;
	int32 type = 4; // 1:同步调用 2:异步调用 3:带超时
	int64 timedOutNanos = 5; // 超时时间:纳秒
}