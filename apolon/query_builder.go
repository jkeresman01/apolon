package apolon

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jkeresman01/apolon/apolon-shared"
)

// Query represents a SELECT query builder
type Query[T any] struct {
	apolon     *DB
	table      string
	columns    []string
	conditions []shared.Condition
	orderBys   []shared.OrderBy
	limit      *int
	offset     *int
	tracking   bool // whether to track returned entities
}

// newQuery creates a new query for the given type
func newQuery[T any](apolon *DB) *Query[T] {
	var zero T
	info := shared.ParseModel(&zero)
	return &Query[T]{
		apolon:   apolon,
		table:    info.Table,
		columns:  info.Fields,
		tracking: true, // tracking enabled by default
	}
}

// Where adds a condition to the query
func (q *Query[T]) Where(c shared.Condition) *Query[T] {
	q.conditions = append(q.conditions, c)
	return q
}

// OrderBy adds an ORDER BY clause to the query
func (q *Query[T]) OrderBy(o shared.OrderBy) *Query[T] {
	q.orderBys = append(q.orderBys, o)
	return q
}

// Limit sets the maximum number of results
func (q *Query[T]) Limit(n int) *Query[T] {
	q.limit = &n
	return q
}

// Offset sets the number of results to skip
func (q *Query[T]) Offset(n int) *Query[T] {
	q.offset = &n
	return q
}

// AsTracking enables change tracking for returned entities (default)
func (q *Query[T]) AsTracking() *Query[T] {
	q.tracking = true
	return q
}

// AsNoTracking disables change tracking for returned entities
// Use this for read-only queries to improve performance
func (q *Query[T]) AsNoTracking() *Query[T] {
	q.tracking = false
	return q
}

// buildSQL constructs the SQL query and arguments
func (q *Query[T]) buildSQL() (string, []any) {
	var sb strings.Builder
	args := []any{}
	paramIdx := 1

	// SELECT
	sb.WriteString("SELECT ")
	sb.WriteString(strings.Join(q.columns, ", "))
	sb.WriteString(" FROM ")
	sb.WriteString(q.table)

	// WHERE
	if len(q.conditions) > 0 {
		sb.WriteString(" WHERE ")
		whereParts := make([]string, 0, len(q.conditions))
		for _, cond := range q.conditions {
			sql, condArgs, nextIdx := cond.ToSQL(paramIdx)
			whereParts = append(whereParts, sql)
			args = append(args, condArgs...)
			paramIdx = nextIdx
		}
		sb.WriteString(strings.Join(whereParts, " AND "))
	}

	// ORDER BY
	if len(q.orderBys) > 0 {
		sb.WriteString(" ORDER BY ")
		orderParts := make([]string, 0, len(q.orderBys))
		for _, o := range q.orderBys {
			orderParts = append(orderParts, o.ToSQL())
		}
		sb.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT
	if q.limit != nil {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", *q.limit))
	}

	// OFFSET
	if q.offset != nil {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", *q.offset))
	}

	return sb.String(), args
}

// ToSQL returns the SQL query string and arguments (for debugging)
func (q *Query[T]) ToSQL() (string, []any) {
	return q.buildSQL()
}

// ToSlice executes the query and returns all matching results
func (q *Query[T]) ToSlice() ([]T, error) {
	sql, args := q.buildSQL()

	rows, err := q.apolon.conn.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var item T
		if err := scanStruct(rows, &item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Track entities if tracking is enabled
	if q.tracking && q.apolon.ChangeTracker != nil {
		for i := range results {
			q.apolon.ChangeTracker.Track(&results[i], shared.Unchanged)
		}
	}

	return results, nil
}

// First returns the first matching result or nil if none found
func (q *Query[T]) First() (*T, error) {
	one := 1
	q.limit = &one
	results, err := q.ToSlice()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// Find finds an entity by its primary key
func (q *Query[T]) Find(pk any) (*T, error) {
	// First check if entity is already tracked
	if q.apolon.ChangeTracker != nil {
		var zero T
		t := reflect.TypeOf(zero)
		if entry := q.apolon.ChangeTracker.GetEntryByKey(t, pk); entry != nil {
			if result, ok := entry.Entity.(*T); ok {
				return result, nil
			}
		}
	}

	// Get PK column name
	var zero T
	pkColName := getPKColumnFromType(zero)

	// Query from database
	q.conditions = append(q.conditions, &shared.SimpleCondition{
		Column: pkColName,
		Op:     "=",
		Value:  pk,
	})

	return q.First()
}

// Count returns the number of matching rows
func (q *Query[T]) Count() (int, error) {
	var sb strings.Builder
	args := []any{}
	paramIdx := 1

	sb.WriteString("SELECT COUNT(*) FROM ")
	sb.WriteString(q.table)

	if len(q.conditions) > 0 {
		sb.WriteString(" WHERE ")
		whereParts := make([]string, 0, len(q.conditions))
		for _, cond := range q.conditions {
			sql, condArgs, nextIdx := cond.ToSQL(paramIdx)
			whereParts = append(whereParts, sql)
			args = append(args, condArgs...)
			paramIdx = nextIdx
		}
		sb.WriteString(strings.Join(whereParts, " AND "))
	}

	var count int
	err := q.apolon.conn.QueryRow(sb.String(), args...).Scan(&count)
	return count, err
}

// Exists returns true if any matching rows exist
func (q *Query[T]) Exists() (bool, error) {
	count, err := q.Count()
	return count > 0, err
}

// Scanner interface for rows.Scan
type scanner interface {
	Scan(dest ...any) error
}

// scanStruct scans a row into a struct
func scanStruct(rows scanner, dest any) error {
	v := reflect.ValueOf(dest).Elem()
	t := v.Type()

	ptrs := make([]any, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		ptrs[i] = v.Field(i).Addr().Interface()
	}

	return rows.Scan(ptrs...)
}

// getPKColumnFromType extracts the PK column name from a type
func getPKColumnFromType(entity any) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Look for field with pk tag
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("apolon")
		if containsPK(tag) {
			parts := strings.Split(tag, ",")
			return parts[0]
		}
	}

	// Default to "id"
	if field, ok := t.FieldByName("ID"); ok {
		tag := field.Tag.Get("apolon")
		if tag != "" {
			parts := strings.Split(tag, ",")
			return parts[0]
		}
		return "id"
	}

	return "id"
}
