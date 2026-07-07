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
	if !strings.Contains(sql, "CREATE TABLE note") {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if !strings.Contains(sql, "body TEXT NOT NULL") {
		t.Fatalf("missing body column: %s", sql)
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
