package migration

import (
	"strings"

	"github.com/lrndwy/gokil/orm"
)

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func resolveFKColumn(meta orm.ModelMeta, fkRef string) string {
	if fm, ok := meta.FieldByName[fkRef]; ok {
		return fm.Column
	}
	return orm.ToColumnName(fkRef)
}
