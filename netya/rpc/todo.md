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

func (hs *HelloService) Hello(req HelloRequest, resp *HelloResponse) error{
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
RpcClient.CallTimedOut("HelloService.Hello", req, resp, timedOutNanos, failRetryCount) error


==============
rpc.proto
package rpc
message RpcCall {
	int64  seq 			  = 1; // 序列号-客户端使用
	string serviceName    = 2; // 服务名， 结构.方法：HelloService.Hello
	bytes  req 			  = 3; //
	bytes  resp 		  = 4; // 
	string err			  = 5; // 错误信息
	bool   reply		  = 6; // 是否需要回复
}