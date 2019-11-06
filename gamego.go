// gamego project gamego.go
package main

import (
	"time"

	"gamego/conf"
	"gamego/dao/db"

	log "github.com/zone1996/logo"
)

func main() {
	conf.Init("./conf/config.conf")

	logconfig := &log.LogConfig{
		Level:        log.LEVEL_INFO,
		SkipFileName: true,
	}
	log.Init(logconfig)

	dbconfig := &db.DbConfig{
		Url: conf.GetConfig().DbConfig["url"].(string),
	}

	initComponent(db.Init(dbconfig), "数据库")

	time.Sleep(time.Second * 1)
}

func initComponent(ok bool, compName string) {
	if !ok {
		log.Fatal(compName + " Init failed")
	}
}
