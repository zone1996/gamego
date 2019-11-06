package entity

const (
	NONE   = 0 // no data changed, need not do anything
	INSERT = 1 // indicates a record need to insert
	UPDATE = 2 // indicates a record need to update
)

type BaseEntity struct {
	option int `gorm:"-"` // 使gorm忽略对此字段的处理
}

func (this *BaseEntity) SetUpdate() {
	if this.IsInsert() {
		return
	}
	this.option = UPDATE
}

func (this *BaseEntity) SetInsert() {
	this.option = INSERT
}

func (this *BaseEntity) SetNone() {
	this.option = NONE
}

func (this *BaseEntity) IsUpdate() bool {
	return this.option == UPDATE
}

func (this *BaseEntity) IsInsert() bool {
	return this.option == INSERT
}
