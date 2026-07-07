package models

import "github.com/lrndwy/gokil/orm"

func init() {
	_ = orm.RegisterModels(
		&User{},
		&Post{},
		&Tag{},
	)
}

type User struct {
	orm.BaseModel
	Email string `orm:"unique,required,size:255"`
	Name  string `orm:"size:100"`
	Posts []Post
}

type Post struct {
	orm.BaseModel
	Title    string `orm:"required,size:200"`
	Content  string `orm:"text"`
	AuthorID int64  `orm:"required"`
	Author   *User
	Tags     []Tag `orm:"many_many:post_tags"`
}

type Tag struct {
	orm.BaseModel
	Name string `orm:"unique,required,size:50"`
}
