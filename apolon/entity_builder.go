package apolon

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jkeresman01/apolon/apolon-shared"
)

// Insert inserts a model into the database immediately (bypassing change tracker)
// For change-tracked inserts, use apolon.Add() followed by apolon.SaveChanges()
func (apolon *DB) Insert(model interface{}) error {
	info := shared.ParseModel(model)
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	cols := []string{}
	vals := []interface{}{}
	placeholders := []string{}
	pkFieldName := ""
	pkColName := ""

	// Find PK field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("apolon")
		if containsPK(tag) {
			pkFieldName = field.Name
			parts := strings.Split(tag, ",")
			pkColName = parts[0]
			break
		}
	}

	// If no PK tag, default to ID field
	if pkFieldName == "" {
		if _, ok := t.FieldByName("ID"); ok {
			pkFieldName = "ID"
			pkColName = "id"
		}
	}

	for i, f := range info.Fields {
		fieldName := t.Field(i).Name

		// Skip auto-increment PK (if value is zero)
		if fieldName == pkFieldName {
			pkVal := v.Field(i).Interface()
			if isZeroValue(pkVal) {
				continue
			}
		}

		cols = append(cols, f)
		vals = append(vals, v.Field(i).Interface())
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(vals)))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		info.Table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	// If PK was skipped (auto-increment), use RETURNING
	if pkFieldName != "" && pkColName != "" {
		// Check if PK was skipped
		pkSkipped := true
		for _, col := range cols {
			if col == pkColName {
				pkSkipped = false
				break
			}
		}

		if pkSkipped {
			query += fmt.Sprintf(" RETURNING %s", pkColName)
			var newPK interface{}
			err := apolon.conn.QueryRow(query, vals...).Scan(&newPK)
			if err != nil {
				return fmt.Errorf("insert failed: %w", err)
			}
			// Set the new PK value on the entity
			setPKValue(model, pkFieldName, newPK)

			// Track as Unchanged after insert
			if apolon.ChangeTracker != nil {
				apolon.ChangeTracker.Track(model, shared.Unchanged)
			}
			return nil
		}
	}

	_, err := apolon.conn.Exec(query, vals...)
	if err != nil {
		return err
	}

	// Track as Unchanged after insert
	if apolon.ChangeTracker != nil {
		apolon.ChangeTracker.Track(model, shared.Unchanged)
	}

	return nil
}
