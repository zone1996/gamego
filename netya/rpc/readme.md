跨语言的RPC调用。

Example:
```protobuffer
syntax = "proto3";

package main;

message HelloRequest {
	string name = 1;
}

message HelloResponse {
	string msg = 1;
}

```

## Server端
```go
type HelloService struct{}

func (hs *HelloService) Hello(req HelloRequest, resp *HelloResponse) error {
	fmt.Println("req name:", req.Name)
	if resp == nil {
		fmt.Println("resp is nil")
		return nil
	}
	resp.Msg = "Hello," + req.Name
	return nil
}

func (hs *HelloService) Hello2(req HelloRequest, resp *HelloResponse) error {
	fmt.Println("req2 name:", req.Name)
	if resp == nil {
		fmt.Println("resp2 is nil")
		return nil
	}
	resp.Msg = "Hello2," + req.Name
	return nil
}


rpcconf := &netya.AcceptorConfig{
	Addr: ":6667",
}
rpcserver := rpc.NewRpcServer(rpcconf)
err := rpcserver.Register(&HelloService{})
if err != nil {
	log.Info("RpcServer register failed,err:?", err)
	return
}
rpcserver.Start()
```

## Client端：
```go
client := rpc.NewRpcClient("localhost:6667")

for i := 1; i <= 10; i++ {
	fmt.Println("===============", i)
	req := &HelloRequest{}
	req.Name = fmt.Sprintf("TestClient%d", i)
	resp := &HelloResponse{}
	err := client.Call("HelloService.Hello", req, resp)
	fmt.Println("err:", err)
	fmt.Println(resp.GetMsg())
}
```