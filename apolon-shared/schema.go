package shared

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ColumnInfo contains metadata about a database column
type ColumnInfo struct {
	Name         string
	GoType       string
	SQLType      string
	IsPrimaryKey bool
	IsNotNull    bool
	IsUnique     bool
	DefaultValue *string
	Size         int
}

// SchemaInfo contains metadata about a database table schema
type SchemaInfo struct {
	Table   string
	Columns []ColumnInfo
}

// ParseSchema extracts schema metadata from a struct using reflection
func ParseSchema(v interface{}) *SchemaInfo {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	columns := []ColumnInfo{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("apolon")

		// Skip unexported fields
		if !f.IsExported() {
			continue
		}

		// Skip fields with apolon:"-"
		if tag == "-" {
			continue
		}

		col := parseColumnInfo(f, tag)
		columns = append(columns, col)
	}

	return &SchemaInfo{
		Table:   strings.ToLower(t.Name()) + "s",
		Columns: columns,
	}
}

// parseColumnInfo parses a struct field into column metadata
func parseColumnInfo(f reflect.StructField, tag string) ColumnInfo {
	col := ColumnInfo{
		GoType: f.Type.String(),
	}

	// Parse tag options
	if tag == "" {
		col.Name = strings.ToLower(f.Name)
	} else {
		parts := strings.Split(tag, ",")
		col.Name = parts[0]

		for _, opt := range parts[1:] {
			parseTagOption(&col, opt)
		}
	}

	// Determine SQL type if not overridden
	if col.SQLType == "" {
		col.SQLType = goTypeToSQLType(f.Type, &col)
	}

	return col
}

// parseTagOption parses a single tag option and updates the column info
func parseTagOption(col *ColumnInfo, opt string) {
	opt = strings.TrimSpace(opt)

	switch {
	case opt == "pk":
		col.IsPrimaryKey = true
		// PKs are implicitly NOT NULL and UNIQUE
		col.IsNotNull = true
		col.IsUnique = true
	case opt == "notnull":
		col.IsNotNull = true
	case opt == "unique":
		col.IsUnique = true
	case strings.HasPrefix(opt, "default:"):
		val := strings.TrimPrefix(opt, "default:")
		col.DefaultValue = &val
	case strings.HasPrefix(opt, "size:"):
		sizeStr := strings.TrimPrefix(opt, "size:")
		if size, err := strconv.Atoi(sizeStr); err == nil {
			col.Size = size
		}
	case strings.HasPrefix(opt, "type:"):
		col.SQLType = strings.TrimPrefix(opt, "type:")
	}
}

// goTypeToSQLType maps Go types to PostgreSQL types
func goTypeToSQLType(t reflect.Type, col *ColumnInfo) string {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check for time.Time
	if t == reflect.TypeOf(time.Time{}) {
		return "TIMESTAMP WITH TIME ZONE"
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32:
		if col.IsPrimaryKey {
			return "SERIAL"
		}
		return "INTEGER"
	case reflect.Int64, reflect.Uint64:
		if col.IsPrimaryKey {
			return "BIGSERIAL"
		}
		return "BIGINT"
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		return "SMALLINT"
	case reflect.String:
		if col.Size > 0 {
			return "VARCHAR(" + strconv.Itoa(col.Size) + ")"
		}
		return "TEXT"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Float32, reflect.Float64:
		return "DOUBLE PRECISION"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "BYTEA"
		}
		return "JSONB"
	default:
		return "TEXT"
	}
}
