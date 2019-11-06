这里是说明Dao结构及使用的简单示例。
## BaseEntity
	定义了一个实体结构应该具有的基本操作

## PlayerInfo
	组合BaseEntity，对应数据库中的一张表；具有一系列Get/Set方法获取/修改字段的值

## PlayerInfoDao
	封装一系列方法和数据库交互，如获取一行记录并返回PlayerInfo结构

## db
	提供一个上层业务使用数据库的顶级接口
	用户需先为daoUtil结构增加xxxDao，然后在register中注册之。
	db.Init后，就可以使用数据库了。
	`
	import "gamego/dao/db"
	db.Dao.PlayerInfo.GetPlayerInfo(userid)
	`
