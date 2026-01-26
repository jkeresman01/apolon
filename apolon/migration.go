package apolon

import (
	"fmt"
	"strings"

	shared "github.com/jkeresman01/apolon/apolon-shared"
)

// MigrationBuilder generates DDL statements for database migrations
type MigrationBuilder struct {
	db *DB
}

// newMigrationBuilder creates a new MigrationBuilder
func newMigrationBuilder(db *DB) *MigrationBuilder {
	return &MigrationBuilder{db: db}
}

// BuildCreateTableSQL generates a CREATE TABLE IF NOT EXISTS statement
func (mb *MigrationBuilder) BuildCreateTableSQL(schema *shared.SchemaInfo) string {
	var sb strings.Builder

	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(schema.Table)
	sb.WriteString(" (\n")

	columnDefs := make([]string, 0, len(schema.Columns))
	for _, col := range schema.Columns {
		columnDefs = append(columnDefs, "    "+mb.buildColumnDefinition(col))
	}

	sb.WriteString(strings.Join(columnDefs, ",\n"))
	sb.WriteString("\n)")

	return sb.String()
}

// buildColumnDefinition generates the SQL for a single column definition
func (mb *MigrationBuilder) buildColumnDefinition(col shared.ColumnInfo) string {
	var parts []string

	// Column name
	parts = append(parts, col.Name)

	// SQL type
	parts = append(parts, col.SQLType)

	// PRIMARY KEY constraint
	if col.IsPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	// NOT NULL constraint (skip for PKs, they're implicitly NOT NULL)
	if col.IsNotNull && !col.IsPrimaryKey {
		parts = append(parts, "NOT NULL")
	}

	// UNIQUE constraint
	if col.IsUnique {
		parts = append(parts, "UNIQUE")
	}

	// DEFAULT value
	if col.DefaultValue != nil {
		parts = append(parts, "DEFAULT", *col.DefaultValue)
	}

	return strings.Join(parts, " ")
}

// AutoMigrate creates tables for the given entities if they don't exist
func (db *DB) AutoMigrate(entities ...any) error {
	mb := newMigrationBuilder(db)

	for _, entity := range entities {
		schema := shared.ParseSchema(entity)
		sql := mb.BuildCreateTableSQL(schema)

		_, err := db.conn.Exec(sql)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", schema.Table, err)
		}
	}

	return nil
}
