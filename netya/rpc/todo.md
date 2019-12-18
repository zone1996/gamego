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
=======================================================================
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

RpcClient.Call("HelloService.Hello", req, resp) error
RpcClient.CallAsync("HelloService.Hello", req) (resp <-chan []byte, err error)
RpcClient.CallTimedOut("HelloService.Hello", req, resp, timedOutNanos) error


==============
rpc.proto
package rpcpb
message RpcCall {
	int64  seq 			 = 1; // 序列号
	string serviceName   = 2; 
	bytes  req 			 = 3;
	bytes  resp 		 = 4;
	int32  type 		 = 5; // 1:同步调用 2:异步调用 3:带超时
	int64  timedOutNanos = 6; // 超时时间:纳秒
}