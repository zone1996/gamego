package db

import (
	"sync"

	log "github.com/zone1996/logo"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DbConfig struct {
	Url string // eg. user:password@tcp(localhost:3306)/dbname?charset=utf8&parseTime=True&loc=Local
}

type basedao struct{}

func (bd *basedao) ExecNoneQuery(sqlFormat string, args ...interface{}) {
	Dao.db.Exec(sqlFormat, args...)
}

type daoUtil struct {
	once          sync.Once
	db            *gorm.DB
	Playerinfodao *PlayerInfoDao
}

var Dao *daoUtil

func Init(dbconfig *DbConfig) (result bool) {
	result = false
	db, err := gorm.Open("mysql", dbconfig.Url)
	if err == nil {
		db.SingularTable(false)
		Dao = &daoUtil{db: db}
		Dao.register()
		result = true
	} else {
		log.Error("Error:?", err)
	}
	return
}

func (this *daoUtil) register() {
	this.Playerinfodao = &PlayerInfoDao{}
}

func (this *daoUtil) Shutdown() {
	this.once.Do(func() {
		if this.db != nil {
			this.db.Close()
		}
	})

	log.Info("数据库模块关闭")
}
