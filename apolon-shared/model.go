package shared

import (
	"reflect"
	"strings"
)

// ModelInfo contains metadata about a database model
type ModelInfo struct {
	Table  string
	Fields []string
}

// ParseModel extracts model metadata using reflection
func ParseModel(v interface{}) *ModelInfo {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := []string{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("apolon")
		if tag == "" {
			tag = strings.ToLower(f.Name)
		} else {
			// Handle comma-separated options like "id,pk" - take only the column name
			if idx := strings.Index(tag, ","); idx != -1 {
				tag = tag[:idx]
			}
		}
		fields = append(fields, tag)
	}

	return &ModelInfo{
		Table:  strings.ToLower(t.Name()) + "s",
		Fields: fields,
	}
}
