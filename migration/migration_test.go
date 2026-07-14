package migration_test

import (
	"strings"
	"testing"

	"github.com/lrndwy/gokil/migration"
	"github.com/lrndwy/gokil/orm"
)

func TestRenderCreateTable(t *testing.T) {
	orm.ResetRegistry()

	type Note struct {
		orm.BaseModel
		Body string `orm:"type:text;not null"`
	}
	_ = orm.RegisterModels(&Note{})
	meta, _ := orm.GetModel("Note")

	sql := migration.RenderCreateTable(*meta)
	if !strings.Contains(sql, `CREATE TABLE "note"`) {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if !strings.Contains(sql, `"body" TEXT NOT NULL`) {
		t.Fatalf("missing body column: %s", sql)
	}
}

func TestRenderUserTableQuoted(t *testing.T) {
	orm.ResetRegistry()

	type User struct {
		orm.BaseModel
		Email string `orm:"unique;not null;size:255"`
	}
	type Post struct {
		orm.BaseModel
		Title    string `orm:"not null;size:200"`
		AuthorID int64  `orm:"not null"`
		Author   *User  `orm:"fk:AuthorID;rel:belongs_to"`
	}
	_ = orm.RegisterModels(&User{}, &Post{})

	userMeta, _ := orm.GetModel("User")
	postMeta, _ := orm.GetModel("Post")

	userSQL := migration.RenderCreateTable(*userMeta)
	if !strings.Contains(userSQL, `CREATE TABLE "user"`) {
		t.Fatalf("user table should be quoted: %s", userSQL)
	}

	postSQL := migration.RenderCreateTable(*postMeta)
	if !strings.Contains(postSQL, `"author_id"`) {
		t.Fatalf("expected author_id column: %s", postSQL)
	}
	if strings.Contains(postSQL, "author_i_d") {
		t.Fatalf("unexpected broken column name: %s", postSQL)
	}
	if strings.Contains(postSQL, "FOREIGN KEY") {
		t.Fatalf("FK should not be inline in CREATE TABLE: %s", postSQL)
	}

	fkSQL := migration.RenderFKConstraint(*postMeta)
	if !strings.Contains(fkSQL, `FOREIGN KEY ("author_id") REFERENCES "user"("id")`) {
		t.Fatalf("unexpected fk sql: %s", fkSQL)
	}
}

func TestExtractSection(t *testing.T) {
	content := "-- +migrate Up\nCREATE TABLE x;\n-- +migrate Down\nDROP TABLE x;"
	up := migration.ExtractSection(content, "Up")
	if !strings.Contains(up, "CREATE TABLE x") {
		t.Fatalf("up section wrong: %s", up)
	}
	down := migration.ExtractSection(content, "Down")
	if !strings.Contains(down, "DROP TABLE x") {
		t.Fatalf("down section wrong: %s", down)
	}
}
