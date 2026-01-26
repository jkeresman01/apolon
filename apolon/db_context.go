package apolon

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/jkeresman01/apolon/apolon-shared"
	_ "github.com/lib/pq"
)

// DB wraps a database connection and provides change tracking
type DB struct {
	conn          *sql.DB
	ChangeTracker *ChangeTracker
}

// Open creates a new database connection with change tracking enabled
func Open(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &DB{
		conn:          conn,
		ChangeTracker: newChangeTracker(),
	}, nil
}

// Close closes the database connection
func (apolon *DB) Close() error {
	return apolon.conn.Close()
}

// Conn returns the underlying sql.DB connection
func (apolon *DB) Conn() *sql.DB {
	return apolon.conn
}

// Set returns a DbSet for the given entity type, providing a fluent query API
func Set[T any](apolon *DB) *DbSet[T] {
	return newDbSet[T](apolon)
}

// Add marks an entity to be inserted on SaveChanges
func (apolon *DB) Add(entity any) *EntityEntry {
	return apolon.ChangeTracker.Track(entity, shared.Added)
}

// AddRange marks multiple entities to be inserted on SaveChanges
func (apolon *DB) AddRange(entities ...any) {
	for _, entity := range entities {
		apolon.Add(entity)
	}
}

// Update marks an entity to be updated on SaveChanges
func (apolon *DB) Update(entity any) *EntityEntry {
	entry := apolon.ChangeTracker.GetEntry(entity)
	if entry != nil {
		entry.State = shared.Modified
		return entry
	}
	return apolon.ChangeTracker.Track(entity, shared.Modified)
}

// Remove marks an entity to be deleted on SaveChanges
func (apolon *DB) Remove(entity any) *EntityEntry {
	entry := apolon.ChangeTracker.GetEntry(entity)
	if entry != nil {
		entry.State = shared.Deleted
		return entry
	}
	return apolon.ChangeTracker.Track(entity, shared.Deleted)
}

// RemoveRange marks multiple entities to be deleted on SaveChanges
func (apolon *DB) RemoveRange(entities ...any) {
	for _, entity := range entities {
		apolon.Remove(entity)
	}
}

// Attach begins tracking an entity as Unchanged
func (apolon *DB) Attach(entity any) *EntityEntry {
	return apolon.ChangeTracker.Track(entity, shared.Unchanged)
}

// Entry returns the EntityEntry for a tracked entity
func (apolon *DB) Entry(entity any) *EntityEntry {
	return apolon.ChangeTracker.GetEntry(entity)
}

// SaveChanges persists all tracked changes to the database
func (apolon *DB) SaveChanges() (int, error) {
	return apolon.SaveChangesContext(nil)
}

// SaveChangesContext persists all tracked changes within an optional transaction
func (apolon *DB) SaveChangesContext(tx *sql.Tx) (int, error) {
	apolon.ChangeTracker.DetectChanges()

	var execer interface {
		Exec(query string, args ...any) (sql.Result, error)
	}

	ownTx := false
	if tx == nil {
		var err error
		tx, err = apolon.conn.Begin()
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}
		ownTx = true
		defer func() {
			if ownTx {
				tx.Rollback()
			}
		}()
	}
	execer = tx

	affected := 0

	// Process Deleted entities first (to avoid FK issues)
	for _, entry := range apolon.ChangeTracker.EntriesByState(shared.Deleted) {
		n, err := apolon.executeDelete(execer, entry)
		if err != nil {
			return affected, err
		}
		affected += n
	}

	// Process Added entities
	for _, entry := range apolon.ChangeTracker.EntriesByState(shared.Added) {
		n, err := apolon.executeInsert(execer, entry)
		if err != nil {
			return affected, err
		}
		affected += n
	}

	// Process Modified entities
	for _, entry := range apolon.ChangeTracker.EntriesByState(shared.Modified) {
		n, err := apolon.executeUpdate(execer, entry)
		if err != nil {
			return affected, err
		}
		affected += n
	}

	// Also check Unchanged entities that may have changes
	for _, entry := range apolon.ChangeTracker.EntriesByState(shared.Unchanged) {
		if entry.HasChanges() {
			n, err := apolon.executeUpdate(execer, entry)
			if err != nil {
				return affected, err
			}
			affected += n
		}
	}

	if ownTx {
		if err := tx.Commit(); err != nil {
			return affected, fmt.Errorf("failed to commit transaction: %w", err)
		}
		ownTx = false
	}

	apolon.ChangeTracker.AcceptAllChanges()
	return affected, nil
}

// executeInsert generates and executes an INSERT statement
func (apolon *DB) executeInsert(execer interface {
	Exec(string, ...any) (sql.Result, error)
}, entry *EntityEntry) (int, error) {
	info := shared.ParseModel(entry.Entity)
	v := reflect.ValueOf(entry.Entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	cols := []string{}
	vals := []any{}
	placeholders := []string{}

	for i, col := range info.Fields {
		// Skip auto-increment PK (if value is zero)
		if entry.pkField != "" && entry.entityType.Field(i).Name == entry.pkField {
			pkVal := v.Field(i).Interface()
			if isZeroValue(pkVal) {
				continue
			}
		}
		cols = append(cols, col)
		vals = append(vals, v.Field(i).Interface())
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(vals)))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		info.Table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	// If PK was skipped, use RETURNING to get the generated value
	if entry.pkField != "" {
		pkColName := getPKColumnName(entry.Entity, entry.pkField)
		if pkColName != "" {
			query += fmt.Sprintf(" RETURNING %s", pkColName)
			var newPK any
			err := apolon.conn.QueryRow(query, vals...).Scan(&newPK)
			if err != nil {
				return 0, fmt.Errorf("insert failed: %w", err)
			}
			// Set the new PK value on the entity
			setPKValue(entry.Entity, entry.pkField, newPK)
			return 1, nil
		}
	}

	result, err := execer.Exec(query, vals...)
	if err != nil {
		return 0, fmt.Errorf("insert failed: %w", err)
	}

	n, _ := result.RowsAffected()
	return int(n), nil
}

// executeUpdate generates and executes an UPDATE statement
func (apolon *DB) executeUpdate(execer interface {
	Exec(string, ...any) (sql.Result, error)
}, entry *EntityEntry) (int, error) {
	changed := entry.GetChangedProperties()
	if len(changed) == 0 {
		return 0, nil
	}

	info := shared.ParseModel(entry.Entity)
	v := reflect.ValueOf(entry.Entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Build SET clause with only changed columns
	setClauses := []string{}
	vals := []any{}
	paramIdx := 1

	for i, col := range info.Fields {
		fieldName := entry.entityType.Field(i).Name
		if _, isChanged := changed[fieldName]; isChanged {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, paramIdx))
			vals = append(vals, v.Field(i).Interface())
			paramIdx++
		}
	}

	// Add PK to WHERE clause
	pkColName := getPKColumnName(entry.Entity, entry.pkField)
	pkValue := entry.GetPrimaryKey()
	vals = append(vals, pkValue)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d",
		info.Table,
		strings.Join(setClauses, ", "),
		pkColName,
		paramIdx,
	)

	result, err := execer.Exec(query, vals...)
	if err != nil {
		return 0, fmt.Errorf("update failed: %w", err)
	}

	n, _ := result.RowsAffected()
	return int(n), nil
}

// executeDelete generates and executes a DELETE statement
func (apolon *DB) executeDelete(execer interface {
	Exec(string, ...any) (sql.Result, error)
}, entry *EntityEntry) (int, error) {
	info := shared.ParseModel(entry.Entity)
	pkColName := getPKColumnName(entry.Entity, entry.pkField)
	pkValue := entry.GetPrimaryKey()

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", info.Table, pkColName)

	result, err := execer.Exec(query, pkValue)
	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	n, _ := result.RowsAffected()
	return int(n), nil
}

// getPKColumnName returns the database column name for the PK field
func getPKColumnName(entity any, pkField string) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, ok := t.FieldByName(pkField)
	if !ok {
		return strings.ToLower(pkField)
	}

	tag := field.Tag.Get("apolon")
	if tag == "" {
		return strings.ToLower(pkField)
	}

	// Get column name (before comma)
	parts := strings.Split(tag, ",")
	return parts[0]
}

// setPKValue sets the primary key value on an entity using reflection
func setPKValue(entity any, pkField string, value any) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(pkField)
	if field.IsValid() && field.CanSet() {
		val := reflect.ValueOf(value)
		if val.Type().ConvertibleTo(field.Type()) {
			field.Set(val.Convert(field.Type()))
		}
	}
}

// isZeroValue checks if a value is the zero value for its type
func isZeroValue(v any) bool {
	return reflect.ValueOf(v).IsZero()
}
