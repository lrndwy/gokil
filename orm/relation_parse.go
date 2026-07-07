package orm

import (
	"reflect"
	"strings"
)

func parseTypedRelation(sf reflect.StructField) (*FieldMeta, bool) {
	typeName := relationTypeName(sf.Type)
	typeStr := sf.Type.String()

	switch typeName {
	case "BelongsTo":
		related := genericTypeArg(typeStr, 0)
		if related == "" {
			return nil, false
		}
		return &FieldMeta{
			Name:        sf.Name,
			Column:      toColumnName(sf.Name),
			GoType:      sf.Type,
			StructField: sf,
			IsRelation:  true,
			Relation: RelationMeta{
				Type:         RelationBelongsTo,
				FKColumn:     sf.Name + "ID",
				RelatedModel: related,
			},
		}, true
	case "HasMany":
		related := genericTypeArg(typeStr, 0)
		if related == "" {
			return nil, false
		}
		return &FieldMeta{
			Name:        sf.Name,
			Column:      toColumnName(sf.Name),
			GoType:      sf.Type,
			StructField: sf,
			IsRelation:  true,
			Relation: RelationMeta{
				Type:         RelationHasMany,
				RelatedModel: related,
			},
		}, true
	case "ManyMany":
		related := genericTypeArg(typeStr, 0)
		if related == "" {
			return nil, false
		}
		tableType := genericTypeArg(typeStr, 1)
		through := throughTableFromType(tableType)
		return &FieldMeta{
			Name:        sf.Name,
			Column:      toColumnName(sf.Name),
			GoType:      sf.Type,
			StructField: sf,
			IsRelation:  true,
			Relation: RelationMeta{
				Type:         RelationManyToMany,
				RelatedModel: related,
				ThroughTable: through,
			},
		}, true
	default:
		return nil, false
	}
}

func genericTypeArg(typeStr string, index int) string {
	start := strings.Index(typeStr, "[")
	end := strings.LastIndex(typeStr, "]")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	inner := typeStr[start+1 : end]
	parts := splitGenericArgs(inner)
	if index >= len(parts) {
		return ""
	}
	return modelNameFromType(parts[index])
}

func splitGenericArgs(inner string) []string {
	parts := []string{}
	current := strings.Builder{}
	depth := 0
	for _, r := range inner {
		switch r {
		case '[', '(', '{':
			depth++
			current.WriteRune(r)
		case ']', ')', '}':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(current.String()))
				current.Reset()
				continue
			}
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	return parts
}

func modelNameFromType(typeName string) string {
	typeName = strings.TrimSpace(typeName)
	typeName = strings.Trim(typeName, `"`)
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		typeName = typeName[idx+1:]
	}
	if idx := strings.LastIndex(typeName, "/"); idx >= 0 {
		typeName = typeName[idx+1:]
	}
	return typeName
}

func relationTypeName(t reflect.Type) string {
	name := t.Name()
	if name != "" {
		if idx := strings.Index(name, "["); idx >= 0 {
			name = name[:idx]
		}
		return name
	}
	s := t.String()
	if idx := strings.LastIndex(s, "."); idx >= 0 {
		s = s[idx+1:]
	}
	if idx := strings.Index(s, "["); idx >= 0 {
		s = s[:idx]
	}
	return s
}

func isBelongsToType(t reflect.Type) bool {
	return relationTypeName(t) == "BelongsTo"
}

func isHasManyType(t reflect.Type) bool {
	return relationTypeName(t) == "HasMany"
}

func belongsToFKValue(v reflect.Value, fkName string) (int64, bool) {
	if fkName == "" {
		return 0, false
	}
	relName := strings.TrimSuffix(fkName, "ID")
	bt := v.FieldByName(relName)
	if !bt.IsValid() || !isBelongsToType(bt.Type()) {
		return 0, false
	}
	id := bt.FieldByName("ID")
	if !id.IsValid() {
		return 0, false
	}
	return id.Int(), true
}

func isManyManyType(t reflect.Type) bool {
	return relationTypeName(t) == "ManyMany"
}

func throughTableFromType(typeName string) string {
	typeName = modelNameFromType(typeName)
	if strings.HasPrefix(typeName, "Table") {
		typeName = strings.TrimPrefix(typeName, "Table")
	}
	if typeName == "" {
		return ""
	}
	return toSnakeCase(typeName)
}
