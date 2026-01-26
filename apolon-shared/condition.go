package shared

import (
	"fmt"
	"strings"
)

// Condition interface - anything that can become a WHERE clause
type Condition interface {
	ToSQL(paramIndex int) (sql string, args []any, nextIndex int)
}

// SimpleCondition represents a basic comparison: column op value
type SimpleCondition struct {
	Column string
	Op     string
	Value  any
}

func (c *SimpleCondition) ToSQL(idx int) (string, []any, int) {
	return fmt.Sprintf("%s %s $%d", c.Column, c.Op, idx), []any{c.Value}, idx + 1
}

// InCondition represents a column IN (values...) clause
type InCondition struct {
	Column string
	Values []any
}

func (c *InCondition) ToSQL(idx int) (string, []any, int) {
	if len(c.Values) == 0 {
		return "FALSE", nil, idx
	}

	placeholders := make([]string, len(c.Values))
	for i := range c.Values {
		placeholders[i] = fmt.Sprintf("$%d", idx+i)
	}

	sql := fmt.Sprintf("%s IN (%s)", c.Column, strings.Join(placeholders, ", "))
	return sql, c.Values, idx + len(c.Values)
}

// BetweenCondition represents a column BETWEEN a AND b clause
type BetweenCondition struct {
	Column string
	Low    any
	High   any
}

func (c *BetweenCondition) ToSQL(idx int) (string, []any, int) {
	sql := fmt.Sprintf("%s BETWEEN $%d AND $%d", c.Column, idx, idx+1)
	return sql, []any{c.Low, c.High}, idx + 2
}

// NullCondition represents IS NULL or IS NOT NULL
type NullCondition struct {
	Column string
	IsNull bool
}

func (c *NullCondition) ToSQL(idx int) (string, []any, int) {
	if c.IsNull {
		return fmt.Sprintf("%s IS NULL", c.Column), nil, idx
	}
	return fmt.Sprintf("%s IS NOT NULL", c.Column), nil, idx
}

// LikeCondition represents a LIKE clause
type LikeCondition struct {
	Column  string
	Pattern string
}

func (c *LikeCondition) ToSQL(idx int) (string, []any, int) {
	return fmt.Sprintf("%s LIKE $%d", c.Column, idx), []any{c.Pattern}, idx + 1
}

// AndCondition combines multiple conditions with AND
type AndCondition struct {
	Conditions []Condition
}

func (c *AndCondition) ToSQL(idx int) (string, []any, int) {
	if len(c.Conditions) == 0 {
		return "TRUE", nil, idx
	}

	parts := make([]string, 0, len(c.Conditions))
	args := []any{}

	for _, cond := range c.Conditions {
		sql, a, nextIdx := cond.ToSQL(idx)
		parts = append(parts, sql)
		args = append(args, a...)
		idx = nextIdx
	}

	return "(" + strings.Join(parts, " AND ") + ")", args, idx
}

// OrCondition combines multiple conditions with OR
type OrCondition struct {
	Conditions []Condition
}

func (c *OrCondition) ToSQL(idx int) (string, []any, int) {
	if len(c.Conditions) == 0 {
		return "FALSE", nil, idx
	}

	parts := make([]string, 0, len(c.Conditions))
	args := []any{}

	for _, cond := range c.Conditions {
		sql, a, nextIdx := cond.ToSQL(idx)
		parts = append(parts, sql)
		args = append(args, a...)
		idx = nextIdx
	}

	return "(" + strings.Join(parts, " OR ") + ")", args, idx
}

// NotCondition negates a condition
type NotCondition struct {
	Condition Condition
}

func (c *NotCondition) ToSQL(idx int) (string, []any, int) {
	sql, args, nextIdx := c.Condition.ToSQL(idx)
	return "NOT (" + sql + ")", args, nextIdx
}

// And combines multiple conditions with AND
func And(conditions ...Condition) Condition {
	return &AndCondition{Conditions: conditions}
}

// Or combines multiple conditions with OR
func Or(conditions ...Condition) Condition {
	return &OrCondition{Conditions: conditions}
}

// Not negates a condition
func Not(condition Condition) Condition {
	return &NotCondition{Condition: condition}
}
