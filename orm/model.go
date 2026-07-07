package orm

import "time"

type BaseModel struct {
	ID        int64     `orm:"pk;auto"`
	CreatedAt time.Time `orm:"auto_now_add"`
	UpdatedAt time.Time `orm:"auto_now"`
}

func (m *BaseModel) GetID() int64 {
	return m.ID
}

func (m *BaseModel) SetID(id int64) {
	m.ID = id
}
