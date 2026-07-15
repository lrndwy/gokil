package orm

import (
	"strings"
	"testing"
)

func TestOnlySelectColumns(t *testing.T) {
	ResetRegistry()

	type User struct {
		BaseModel
		Email string `orm:"unique,required,size:255"`
		Name  string `orm:"size:100"`
		Bio   string `orm:"text"`
	}

	if err := RegisterModels(&User{}); err != nil {
		t.Fatal(err)
	}

	qs := Objects[User](nil)
	query, _, err := qs.Only("Name", "Email").buildSelect()
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{`"id"`, `"name"`, `"email"`} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %s: %s", want, query)
		}
	}
	if strings.Contains(query, `"bio"`) {
		t.Fatalf("query should not include bio: %s", query)
	}
	if strings.Contains(query, `"created_at"`) {
		t.Fatalf("query should not include created_at: %s", query)
	}
}

func TestOnlyUnknownField(t *testing.T) {
	ResetRegistry()

	type Tag struct {
		BaseModel
		Name string `orm:"size:50"`
	}
	if err := RegisterModels(&Tag{}); err != nil {
		t.Fatal(err)
	}

	_, _, err := Objects[Tag](nil).Only("Nope").buildSelect()
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestOnlyBelongsToResolvesFK(t *testing.T) {
	ResetRegistry()

	type Author struct {
		BaseModel
		Name string
	}
	type Post struct {
		BaseModel
		Title  string
		Body   string `orm:"text"`
		Author BelongsTo[Author] `orm:"required"`
	}
	if err := RegisterModels(&Author{}, &Post{}); err != nil {
		t.Fatal(err)
	}

	query, _, err := Objects[Post](nil).Only("Title", "Author").buildSelect()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"id"`, `"title"`, `"author_id"`} {
		if !strings.Contains(query, want) {
			t.Fatalf("query missing %s: %s", want, query)
		}
	}
	if strings.Contains(query, `"body"`) {
		t.Fatalf("query should not include body: %s", query)
	}
}
