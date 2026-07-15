package orm_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/lrndwy/gokil/orm"
)

func TestProjectForJSONOnlyFields(t *testing.T) {
	orm.ResetRegistry()

	type User struct {
		orm.BaseModel
		Email string `orm:"size:255"`
		Name  string `orm:"size:100"`
	}
	if err := orm.RegisterModels(&User{}); err != nil {
		t.Fatal(err)
	}

	u := &User{}
	u.ID = 7
	u.Name = "Hafiz"
	u.Email = "should-not-appear@example.com"

	orm.SetProjection(u, []string{"ID", "Name"})
	projected := orm.ProjectForJSON(u)

	raw, err := json.Marshal(projected)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["Name"] != "Hafiz" {
		t.Fatalf("Name = %v", m["Name"])
	}
	if _, ok := m["Email"]; ok {
		t.Fatalf("Email should be omitted, got %v", m)
	}
	if _, ok := m["CreatedAt"]; ok {
		t.Fatalf("CreatedAt should be omitted, got %v", m)
	}
	if id, ok := m["ID"].(float64); !ok || id != 7 {
		t.Fatalf("ID = %v", m["ID"])
	}
}

func TestProjectForJSONSlice(t *testing.T) {
	type row struct {
		orm.BaseModel
		Name  string
		Email string
	}
	u1 := &row{}
	u1.ID = 1
	u1.Name = "A"
	u1.Email = "a@x"
	orm.SetProjection(u1, []string{"ID", "Name"})

	raw, err := json.Marshal(orm.ProjectForJSON([]*row{u1}))
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, `"Name":"A"`) || !strings.Contains(s, `"ID":1`) {
		t.Fatalf("unexpected json: %s", s)
	}
	if strings.Contains(s, `"Email"`) {
		t.Fatalf("email should be omitted: %s", s)
	}
}
