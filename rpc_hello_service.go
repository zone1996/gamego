package main

import (
	"fmt"
)

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
