package generator

// FieldInfo represents metadata about a struct field
type FieldInfo struct {
	Name      string // Go field name
	Column    string // Database column name
	FieldType string // ORM field type (IntField, StringField, etc.)
	GoType    string // Original Go type
	IsPK      bool   // Is primary key
}

// ModelInfo represents metadata about a model struct
type ModelInfo struct {
	Name          string      // Struct name
	Table         string      // Table name
	Fields        []FieldInfo // Field information
	Package       string      // Package name
	HasTimeImport bool        // Whether time.Time is used
}
