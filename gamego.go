// gamego project gamego.go
package main

import (
	"gamego/cmd"
	"gamego/conf"
	"gamego/dao/db"
	"gamego/netya"
	"os/signal"

	"os"

	log "github.com/zone1996/logo"
)

func main() {
	conf.Init("./conf/config.conf")
	initLog()
	cmd.InitCmd()
	ac := initNet()
	initDB()
	log.Info("Server started.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	onStop(ac)
}

func initComponent(ok bool, compName string) {
	if !ok {
		log.Fatal(compName + " Init failed")
	}
}

func initLog() {
	logconfig := &log.LogConfig{
		Level:        log.LEVEL_INFO,
		SkipFileName: true,
	}
	log.Init(logconfig)
}

func initNet() *netya.Acceptor {
	netConfig := &netya.AcceptorConfig{
		Port: ":6666",
	}
	ac := netya.NewAcceptor(netConfig, &DefaultHandler{}, &netya.DefaultCodec{})
	go ac.Accept()
	return ac
}

func initDB() {
	dbconfig := &db.DbConfig{
		Url: conf.GetConfig().DbConfig["url"].(string),
	}
	initComponent(db.Init(dbconfig), "数据库")
}

func onStop(ac *netya.Acceptor) {
	ac.Shutdown()
	log.Info("Server Close.")
}
