package orm

import (
	"fmt"
	"reflect"
	"strings"
)

// finalizeModelRelations validates and completes relation metadata after all fields are parsed.
func finalizeModelRelations(meta *ModelMeta) error {
	meta.Relations = nil

	for i := range meta.Fields {
		fm := &meta.Fields[i]
		if !fm.IsRelation {
			continue
		}

		switch fm.Relation.Type {
		case RelationBelongsTo:
			if err := finalizeBelongsTo(meta, fm); err != nil {
				return fmt.Errorf("%s.%s: %w", meta.Name, fm.Name, err)
			}
		case RelationHasMany, RelationReverse:
			finalizeHasMany(meta, fm)
		case RelationManyToMany:
			finalizeManyToMany(meta, fm)
		}

		meta.Relations = append(meta.Relations, *fm)
	}

	injectSyntheticFKFields(meta)
	return nil
}

func finalizeBelongsTo(meta *ModelMeta, fm *FieldMeta) error {
	if fm.Relation.FKColumn == "" {
		fm.Relation.FKColumn = fm.Name + "ID"
	}
	if _, ok := meta.FieldByName[fm.Relation.FKColumn]; !ok {
		if !isBelongsToType(fm.GoType) {
			return fmt.Errorf("belongs_to requires field %q", fm.Relation.FKColumn)
		}
	}
	if fm.Relation.RelatedModel == "" {
		fm.Relation.RelatedModel = relatedModelName(fm.GoType)
	}
	return nil
}

func finalizeHasMany(meta *ModelMeta, fm *FieldMeta) {
	if fm.Relation.RelatedModel == "" {
		fm.Relation.RelatedModel = relatedModelName(fm.GoType)
	}

	fk := fm.Relation.FKColumn
	if fk == "" {
		fk = fm.Relation.ReverseField
	}
	if fk != "" {
		fk = resolveFKFieldName(fk, fm.Relation.RelatedModel)
	}

	if fk == "" {
		fk = inferReverseFK(meta.Name, fm.Relation.RelatedModel)
	}

	if fk == "" {
		fm.Relation.Type = RelationManyToMany
		finalizeManyToMany(meta, fm)
		return
	}

	fm.Relation.FKColumn = fk
	fm.Relation.ReverseField = fk
}

func finalizeManyToMany(meta *ModelMeta, fm *FieldMeta) {
	if fm.Relation.RelatedModel == "" {
		fm.Relation.RelatedModel = relatedModelName(fm.GoType)
	}
	if fm.Relation.ThroughTable == "" {
		fm.Relation.ThroughTable = defaultThroughTable(meta.Name, fm.Relation.RelatedModel)
	}
}

func inferReverseFK(parentModel, childModel string) string {
	childMeta, ok := globalRegistry.models[childModel]
	if !ok {
		return ""
	}

	for _, rel := range childMeta.Relations {
		if rel.Relation.Type == RelationBelongsTo && rel.Relation.RelatedModel == parentModel {
			return rel.Relation.FKColumn
		}
	}

	for _, field := range childMeta.Fields {
		if field.IsRelation && field.Relation.Type == RelationBelongsTo && field.Relation.RelatedModel == parentModel {
			return field.Relation.FKColumn
		}
	}

	for _, field := range childMeta.Fields {
		if field.IsRelation || !strings.HasSuffix(field.Name, "ID") {
			continue
		}
		base := strings.TrimSuffix(field.Name, "ID")
		relField, ok := childMeta.FieldByName[base]
		if !ok || !relField.IsRelation {
			continue
		}
		if relField.Relation.RelatedModel == parentModel {
			return field.Name
		}
	}

	return ""
}

func resolveHasManyFKColumn(parentMeta *ModelMeta, relField *FieldMeta, relatedMeta *ModelMeta) string {
	fkField := relField.Relation.FKColumn
	if fkField == "" {
		fkField = relField.Relation.ReverseField
	}
	if fkField != "" {
		fkField = resolveFKFieldName(fkField, relatedMeta.Name)
		if fm, ok := relatedMeta.FieldByName[fkField]; ok {
			return fm.Column
		}
		return toColumnName(fkField)
	}

	if fk := inferReverseFK(parentMeta.Name, relatedMeta.Name); fk != "" {
		if fm, ok := relatedMeta.FieldByName[fk]; ok {
			return fm.Column
		}
		return toColumnName(fk)
	}

	return toColumnName(parentMeta.Name) + "_id"
}

func normalizeFKFieldName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if !strings.HasSuffix(name, "ID") {
		return name + "ID"
	}
	return name
}

func resolveFKFieldName(name, childModel string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if strings.HasSuffix(name, "ID") {
		return name
	}

	childMeta, ok := globalRegistry.models[childModel]
	if ok {
		candidates := []string{
			name,
			strings.ToUpper(name[:1]) + name[1:],
		}
		for _, candidate := range candidates {
			if rel, ok := childMeta.FieldByName[candidate]; ok && rel.IsRelation && rel.Relation.FKColumn != "" {
				return rel.Relation.FKColumn
			}
			if fm, ok := childMeta.FieldByName[candidate+"ID"]; ok {
				return fm.Name
			}
		}
	}

	return normalizeFKFieldName(name)
}

func defaultThroughTable(a, b string) string {
	left := toTableName(a)
	right := toTableName(b)
	if left > right {
		left, right = right, left
	}
	return left + "_" + right
}

func injectSyntheticFKFields(meta *ModelMeta) {
	for i := range meta.Fields {
		fm := &meta.Fields[i]
		if !fm.IsRelation || fm.Relation.Type != RelationBelongsTo || !isBelongsToType(fm.GoType) {
			continue
		}
		fkName := fm.Relation.FKColumn
		if _, ok := meta.FieldByName[fkName]; ok {
			continue
		}
		synthetic := FieldMeta{
			Name:          fkName,
			Column:        toColumnName(fkName),
			GoType:        reflect.TypeOf(int64(0)),
			FieldType:     FieldTypeBigInt,
			Nullable:      fm.Nullable,
			VirtualFK:     true,
			RelationOwner: fm.Name,
		}
		meta.Fields = append(meta.Fields, synthetic)
		idx := len(meta.Fields) - 1
		meta.FieldByName[fkName] = &meta.Fields[idx]
		meta.FieldByColumn[synthetic.Column] = &meta.Fields[idx]
	}
}
