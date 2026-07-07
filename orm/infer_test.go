package orm_test

import (
	"testing"

	"github.com/lrndwy/gokil/orm"
)

type inferTag struct {
	orm.BaseModel
	Name string
}

type inferPost struct {
	orm.BaseModel
	Title    string
	AuthorID int64
	Author   *inferUser
	Tags     []inferTag `orm:"many_many:post_tags"`
}

type inferUser struct {
	orm.BaseModel
	Email string
	Posts []inferPost
}

type inferCategory struct {
	orm.BaseModel
	Name string
}

type inferItem struct {
	orm.BaseModel
	CategoryID int64
	Category   *inferCategory `orm:"belongs_to:CategoryID"`
}

type inferBook struct {
	orm.BaseModel
	Title    string
	AuthorID int64
}

type inferAuthor struct {
	orm.BaseModel
	Name  string
	Books []inferBook `orm:"has_many:AuthorID"`
}

type legacyPost struct {
	orm.BaseModel
	AuthorID int64
	Author   *legacyUser `orm:"fk:AuthorID;rel:belongs_to"`
}

type legacyUser struct {
	orm.BaseModel
	Posts []legacyPost `orm:"reverse:author"`
}

type typedTag struct {
	orm.BaseModel
	Name string
}

type typedUser struct {
	orm.BaseModel
	Email string
	Posts orm.HasMany[typedPost]
}

type TablePostTags string

type typedPost struct {
	orm.BaseModel
	Title  string
	Author orm.BelongsTo[typedUser] `orm:"required"`
	Tags   orm.ManyMany[typedTag, TablePostTags]
}

func TestTypedRelationSyntax(t *testing.T) {
	orm.ResetRegistry()

	if err := orm.RegisterModels(&typedUser{}, &typedPost{}, &typedTag{}); err != nil {
		t.Fatal(err)
	}

	userMeta, _ := orm.GetModel("typedUser")
	postMeta, _ := orm.GetModel("typedPost")

	postsField := userMeta.FieldByName["Posts"]
	if postsField == nil || postsField.Relation.Type != orm.RelationHasMany {
		t.Fatalf("Posts relation = %+v", postsField)
	}
	if postsField.Relation.FKColumn != "AuthorID" {
		t.Fatalf("Posts FK = %q, want AuthorID", postsField.Relation.FKColumn)
	}

	authorField := postMeta.FieldByName["Author"]
	if authorField == nil || authorField.Relation.Type != orm.RelationBelongsTo {
		t.Fatalf("Author relation = %+v", authorField)
	}
	if authorField.Relation.FKColumn != "AuthorID" {
		t.Fatalf("Author FK = %q, want AuthorID", authorField.Relation.FKColumn)
	}

	authorIDField := postMeta.FieldByName["AuthorID"]
	if authorIDField == nil || !authorIDField.VirtualFK {
		t.Fatalf("expected virtual AuthorID field, got %+v", authorIDField)
	}

	tagsField := postMeta.FieldByName["Tags"]
	if tagsField == nil || tagsField.Relation.Type != orm.RelationManyToMany {
		t.Fatalf("Tags relation = %+v", tagsField)
	}
	if tagsField.Relation.ThroughTable != "post_tags" {
		t.Fatalf("Tags through = %q, want post_tags", tagsField.Relation.ThroughTable)
	}
}

func TestRelationSyntaxSimple(t *testing.T) {
	orm.ResetRegistry()

	if err := orm.RegisterModels(&inferUser{}, &inferPost{}, &inferTag{}); err != nil {
		t.Fatal(err)
	}

	userMeta, _ := orm.GetModel("inferUser")
	postMeta, _ := orm.GetModel("inferPost")

	postsField := userMeta.FieldByName["Posts"]
	if postsField == nil || postsField.Relation.Type != orm.RelationHasMany {
		t.Fatalf("Posts relation = %+v", postsField)
	}
	if postsField.Relation.FKColumn != "AuthorID" {
		t.Fatalf("Posts FK = %q, want AuthorID", postsField.Relation.FKColumn)
	}

	authorField := postMeta.FieldByName["Author"]
	if authorField == nil || authorField.Relation.Type != orm.RelationBelongsTo {
		t.Fatalf("Author relation = %+v", authorField)
	}
	if authorField.Relation.FKColumn != "AuthorID" {
		t.Fatalf("Author FK = %q, want AuthorID", authorField.Relation.FKColumn)
	}
}

func TestBelongsToShorthandTag(t *testing.T) {
	orm.ResetRegistry()

	if err := orm.RegisterModels(&inferCategory{}, &inferItem{}); err != nil {
		t.Fatal(err)
	}

	itemMeta, _ := orm.GetModel("inferItem")
	field := itemMeta.FieldByName["Category"]
	if field == nil || field.Relation.FKColumn != "CategoryID" {
		t.Fatalf("unexpected category relation: %+v", field)
	}
}

func TestHasManyShorthandTag(t *testing.T) {
	orm.ResetRegistry()

	if err := orm.RegisterModels(&inferAuthor{}, &inferBook{}); err != nil {
		t.Fatal(err)
	}

	authorMeta, _ := orm.GetModel("inferAuthor")
	field := authorMeta.FieldByName["Books"]
	if field == nil || field.Relation.FKColumn != "AuthorID" {
		t.Fatalf("unexpected books relation: %+v", field)
	}
}

func TestLegacyRelationTagsStillWork(t *testing.T) {
	orm.ResetRegistry()

	if err := orm.RegisterModels(&legacyUser{}, &legacyPost{}); err != nil {
		t.Fatal(err)
	}

	userMeta, _ := orm.GetModel("legacyUser")
	field := userMeta.FieldByName["Posts"]
	if field == nil || field.Relation.FKColumn != "AuthorID" {
		t.Fatalf("legacy reverse relation FK = %+v", field)
	}
}

func TestRequiredFieldAlias(t *testing.T) {
	orm.ResetRegistry()

	type Product struct {
		orm.BaseModel
		SKU string `orm:"required,size:64"`
	}

	if err := orm.RegisterModels(&Product{}); err != nil {
		t.Fatal(err)
	}

	meta, _ := orm.GetModel("Product")
	field := meta.FieldByName["SKU"]
	if field == nil || field.Nullable {
		t.Fatal("expected required SKU field")
	}
}
