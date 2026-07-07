package models

import "gokil/orm"

func init() {
	_ = orm.RegisterModels(
		&User{},
		&Post{},
		&Tag{},
	)
}

type User struct {
	orm.BaseModel
	Email string `orm:"unique;not null;size:255"`
	Name  string `orm:"size:100"`
	Posts []Post `orm:"reverse:author"`
}

type Post struct {
	orm.BaseModel
	Title    string `orm:"not null;size:200"`
	Content  string `orm:"type:text"`
	AuthorID int64  `orm:"not null"`
	Author   *User  `orm:"fk:AuthorID;rel:belongs_to"`
	Tags     []Tag  `orm:"m2m:post_tags"`
}

type Tag struct {
	orm.BaseModel
	Name string `orm:"unique;not null;size:50"`
}
