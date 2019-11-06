package entity

type PlayerInfo struct {
	BaseEntity
	UserID   int32  `gorm:"column:UserID"` // 设置对应的字段
	NickName string `gorm:"column:NickName"`
}

// 自定义此entity对应的表名，否则默认表名为 playerinfos
func (PlayerInfo) TableName() string {
	return "player_info"
}

func (pi *PlayerInfo) SetUserID(userId int32) {
	if pi.UserID != userId {
		pi.UserID = userId
		pi.SetUpdate()
	}
}

func (pi *PlayerInfo) SetNickName(nickname string) {
	if pi.NickName != nickname {
		pi.NickName = nickname
		pi.SetUpdate()
	}
}

func (pi *PlayerInfo) GetUserID() int32 {
	return pi.UserID
}

func (pi *PlayerInfo) GetNickName() string {
	return pi.NickName
}
