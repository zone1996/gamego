// gamego project gamego.go
package main

import (
	"gamego/cmd"
	"gamego/conf"
	"gamego/dao/db"
	"gamego/netya"
	"gamego/netya/rpc"

	"os"
	"os/signal"

	log "github.com/zone1996/logo"
)

func main() {
	// TODO 从命令行解析配置文件路径
	ac := start()

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	onStop(ac)
	rpcserver.Stop()
}

func initComponent(ok bool, compName string) {
	if !ok {
		log.Fatal(compName + " Init failed")
	}
}

func initLog() {
	logconfig := &log.LogConfig{
		Level:        log.LEVEL_DEBUG,
		SkipFileName: true,
		IsConsole:    true,
	}
	log.Init(logconfig)
}

func initNet() netya.Acceptor {
	netConfig := &netya.AcceptorConfig{
		Addr: ":6666",
	}
	h := &DefaultHandler{}
	ac := netya.NewTCPAcceptor(netConfig, h)
	h.acceptor = ac
	go ac.Accept()
	return ac
}

func initDB() {
	dbconfig := &db.DbConfig{
		Url: conf.GetConfig().DbConfig["url"].(string),
	}
	initComponent(db.Init(dbconfig), "数据库")
}

func start() netya.Acceptor {
	conf.Init("./conf/config.conf")
	initLog()
	cmd.InitCmd()
	ac := initNet()
	initDB()
	log.Info("Server started.")
	return ac
}

func onStop(ac netya.Acceptor) {
	ac.Shutdown()
	log.Info("Server Close.")
}
