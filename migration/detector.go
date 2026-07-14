package migration

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lrndwy/gokil/orm"
)

type Detector struct {
	DB *sql.DB
}

type SchemaDiff struct {
	CreateTables []orm.ModelMeta
	AlterTables  []TableAlter
	CreateM2M    []M2MTable
}

type TableAlter struct {
	Table   string
	AddCols []orm.FieldMeta
}

type M2MTable struct {
	Name    string
	Source  string
	Target  string
}

func (d *Detector) Detect() (*SchemaDiff, error) {
	diff := &SchemaDiff{}
	existing, err := d.listTables()
	if err != nil {
		return nil, err
	}

	for _, meta := range orm.AllModels() {
		if _, ok := existing[meta.TableName]; !ok {
			diff.CreateTables = append(diff.CreateTables, *meta)
		} else {
			alter, err := d.detectAlter(meta, existing[meta.TableName])
			if err != nil {
				return nil, err
			}
			if len(alter.AddCols) > 0 {
				diff.AlterTables = append(diff.AlterTables, alter)
			}
		}

		for _, rel := range meta.Relations {
			if rel.Relation.Type == orm.RelationManyToMany {
				through := rel.Relation.ThroughTable
				if through == "" {
					through = meta.TableName + "_" + orm.ToTableName(rel.Relation.RelatedModel)
				}
				if _, ok := existing[through]; !ok {
					diff.CreateM2M = append(diff.CreateM2M, M2MTable{
						Name:   through,
						Source: meta.TableName,
						Target: orm.ToTableName(rel.Relation.RelatedModel),
					})
				}
			}
		}
	}

	return diff, nil
}

func (d *Detector) listTables() (map[string]map[string]bool, error) {
	rows, err := d.DB.Query(`
		SELECT table_name, column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := map[string]map[string]bool{}
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			return nil, err
		}
		if tables[table] == nil {
			tables[table] = map[string]bool{}
		}
		tables[table][column] = true
	}
	return tables, rows.Err()
}

func (d *Detector) detectAlter(meta *orm.ModelMeta, columns map[string]bool) (TableAlter, error) {
	alter := TableAlter{Table: meta.TableName}
	for _, f := range meta.Fields {
		if f.IsRelation {
			continue
		}
		if !columns[f.Column] {
			alter.AddCols = append(alter.AddCols, f)
		}
	}
	return alter, nil
}

func HasChanges(diff *SchemaDiff) bool {
	return len(diff.CreateTables) > 0 || len(diff.AlterTables) > 0 || len(diff.CreateM2M) > 0
}

func RenderDiff(diff *SchemaDiff) (up, down string) {
	var upB, downB strings.Builder
	upB.WriteString("-- +migrate Up\n")
	downB.WriteString("-- +migrate Down\n")

	for _, meta := range diff.CreateTables {
		upB.WriteString(RenderCreateTable(meta))
		upB.WriteString("\n")
		downB.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", quoteIdent(meta.TableName)))
	}

	for _, meta := range diff.CreateTables {
		if fk := RenderFKConstraint(meta); fk != "" {
			upB.WriteString(fk)
		}
	}

	for _, alter := range diff.AlterTables {
		for _, col := range alter.AddCols {
			upB.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;\n",
				quoteIdent(alter.Table), RenderColumn(col)))
			downB.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;\n",
				quoteIdent(alter.Table), quoteIdent(col.Column)))
		}
	}

	for _, m2m := range diff.CreateM2M {
		upB.WriteString(RenderM2MTable(m2m))
		upB.WriteString("\n")
		downB.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", quoteIdent(m2m.Name)))
	}

	return upB.String(), downB.String()
}

func RenderCreateTable(meta orm.ModelMeta) string {
	table := quoteIdent(meta.TableName)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table))
	cols := []string{}
	for _, f := range meta.Fields {
		if f.IsRelation && f.Relation.Type != orm.RelationBelongsTo {
			continue
		}
		if f.IsRelation && f.Relation.Type == orm.RelationBelongsTo {
			continue
		}
		cols = append(cols, "    "+RenderColumn(f))
	}
	b.WriteString(strings.Join(cols, ",\n"))
	b.WriteString("\n);")

	for _, f := range meta.Fields {
		if f.Unique && !f.PrimaryKey {
			b.WriteString(fmt.Sprintf("\nCREATE UNIQUE INDEX IF NOT EXISTS %s_%s_unique ON %s (%s);",
				meta.TableName, f.Column, table, quoteIdent(f.Column)))
		}
		if f.Index {
			b.WriteString(fmt.Sprintf("\nCREATE INDEX IF NOT EXISTS %s_%s_idx ON %s (%s);",
				meta.TableName, f.Column, table, quoteIdent(f.Column)))
		}
	}

	return b.String()
}

func RenderFKConstraint(meta orm.ModelMeta) string {
	table := quoteIdent(meta.TableName)
	var lines []string
	for _, f := range meta.Fields {
		if f.IsRelation && f.Relation.Type == orm.RelationBelongsTo {
			fkCol := resolveFKColumn(meta, f.Relation.FKColumn)
			if fkCol == "" {
				continue
			}
			relatedTable := quoteIdent(orm.ToTableName(f.Relation.RelatedModel))
			constraint := fmt.Sprintf("%s_%s_fk", meta.TableName, fkCol)
			lines = append(lines, fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s);",
				table, quoteIdent(constraint), quoteIdent(fkCol), relatedTable, quoteIdent("id"),
			))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func RenderColumn(f orm.FieldMeta) string {
	parts := []string{quoteIdent(f.Column), orm.SQLType(f)}
	if f.AutoIncrement {
		parts = append(parts, "GENERATED BY DEFAULT AS IDENTITY")
	}
	if f.PrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}
	if f.Unique && !f.PrimaryKey {
		parts = append(parts, "UNIQUE")
	}
	if !f.Nullable && !f.PrimaryKey {
		parts = append(parts, "NOT NULL")
	}
	if strings.TrimSpace(f.Default) != "" {
		parts = append(parts, "DEFAULT "+f.Default)
	}
	return strings.Join(parts, " ")
}

func RenderM2MTable(m2m M2MTable) string {
	srcCol := toSingularID(m2m.Source)
	dstCol := toSingularID(m2m.Target)
	table := quoteIdent(m2m.Name)
	srcTable := quoteIdent(m2m.Source)
	dstTable := quoteIdent(m2m.Target)

	return fmt.Sprintf(`CREATE TABLE %s (
    %s BIGINT NOT NULL REFERENCES %s(%s) ON DELETE CASCADE,
    %s BIGINT NOT NULL REFERENCES %s(%s) ON DELETE CASCADE,
    PRIMARY KEY (%s, %s)
);`, table,
		quoteIdent(srcCol), srcTable, quoteIdent("id"),
		quoteIdent(dstCol), dstTable, quoteIdent("id"),
		quoteIdent(srcCol), quoteIdent(dstCol))
}

func toSingularID(table string) string {
	col := table
	if strings.HasSuffix(col, "ies") {
		col = strings.TrimSuffix(col, "ies") + "y"
	} else if strings.HasSuffix(col, "s") {
		col = strings.TrimSuffix(col, "s")
	}
	return col + "_id"
}
