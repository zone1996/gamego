package db

import (
	"gamego/dao/entity"
)

type PlayerInfoDao struct {
	*basedao
}

func (this *PlayerInfoDao) Insert(playerinfo *entity.PlayerInfo) (err error) {
	if !playerinfo.IsInsert() {
		return
	}

	if err = Dao.db.Create(playerinfo).Error; err == nil {
		playerinfo.SetNone()
	}
	return
}

func (this *PlayerInfoDao) Update(playerinfo *entity.PlayerInfo) (err error) {
	if !playerinfo.IsUpdate() {
		return
	}

	err = Dao.db.Model(entity.PlayerInfo{}).Where("UserId = ?", playerinfo.GetUserID()).Update(playerinfo).Error
	if err == nil {
		playerinfo.SetNone()
	}
	return
}

func (this *PlayerInfoDao) GetPlayerInfo(userId int32) *entity.PlayerInfo {
	pi := &entity.PlayerInfo{}
	if Dao.db.Where("UserID = ?", userId).First(pi).RecordNotFound() {
		return nil
	}
	pi.SetNone()
	return pi
}
