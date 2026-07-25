package orm

import "time"

type BaseModel struct {
	ID        int64     `orm:"pk;auto" json:"id"`
	CreatedAt time.Time `orm:"auto_now_add" json:"created_at"`
	UpdatedAt time.Time `orm:"auto_now" json:"updated_at"`
}

func (m *BaseModel) GetID() int64 {
	return m.ID
}

func (m *BaseModel) SetID(id int64) {
	m.ID = id
}
