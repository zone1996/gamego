package db

import (
	"fmt"
	"gamego/conf"
	"gamego/dao/entity"
	"testing"
)

func TestDb(t *testing.T) {
	err0 := conf.Init("../../conf/config.conf")
	if err0 != nil {
		t.Errorf("配置初始化失败:%v", err0)
	}
	fmt.Println(conf.GetConfig().DbConfig["url"])

	dbconfig := &DbConfig{
		Url: conf.GetConfig().DbConfig["url"].(string),
	}
	result, err := Init(dbconfig)

	if result {

		csql := `
		CREATE TABLE player_info (
		  UserID int(11) NOT NULL COMMENT '玩家ID',
		  NickName varchar(64) NOT NULL DEFAULT '' COMMENT '昵称',
		  PRIMARY KEY (UserID)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;`

		Dao.Playerinfodao.ExecNoneQuery("DROP TABLE IF EXISTS `player_info`")
		Dao.Playerinfodao.ExecNoneQuery(csql)

		info1 := &entity.PlayerInfo{
			UserID:   1,
			NickName: "a",
		}
		info1.SetInsert()
		fmt.Println("Insert info1: ", Dao.Playerinfodao.Insert(info1)) // 插入新的
		info11 := Dao.Playerinfodao.GetPlayerInfo(1)                   // 查询一条已存在的数据
		fmt.Println("info1: ", info11)

		info2 := &entity.PlayerInfo{
			UserID:   2,
			NickName: "b",
		}
		info2.SetInsert()
		Dao.Playerinfodao.Insert(info2)
		info2.SetNickName("b-b")
		fmt.Println("Update info2:", Dao.Playerinfodao.Update(info2)) // 更新
		info22 := Dao.Playerinfodao.GetPlayerInfo(2)
		fmt.Println("info2: ", info22)

		info3 := Dao.Playerinfodao.GetPlayerInfo(3) // 获取一条不存在的记录
		fmt.Println("info3: ", info3)

		// Dao.Playerinfodao.ExecNoneQuery("delete from player_info where userId=?", 2) // 删除一条数据
	} else {
		t.Errorf("数据库初始化失败:%v", err)
	}

}
