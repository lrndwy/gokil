package orm_test

import (
	"testing"

	"github.com/lrndwy/gokil/orm"
)

func TestRegisterModels(t *testing.T) {
	orm.ResetRegistry()

	type Article struct {
		orm.BaseModel
		Title string `orm:"not null;size:200"`
	}

	err := orm.RegisterModels(&Article{})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	meta, ok := orm.GetModel("Article")
	if !ok {
		t.Fatal("article not found")
	}
	if meta.TableName != "article" {
		t.Fatalf("expected table article, got %s", meta.TableName)
	}
}

func TestParseFieldTags(t *testing.T) {
	orm.ResetRegistry()

	type Product struct {
		orm.BaseModel
		SKU   string `orm:"unique;not null;size:64"`
		Price float64
	}

	_ = orm.RegisterModels(&Product{})
	meta, _ := orm.GetModel("Product")

	skuField := meta.FieldByName["SKU"]
	if skuField == nil || !skuField.Unique {
		t.Fatal("expected SKU to be unique")
	}
}

func TestToTableName(t *testing.T) {
	if orm.ToTableName("BlogPost") != "blog_post" {
		t.Fatal("snake case conversion failed")
	}
}
