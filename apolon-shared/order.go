package shared

import "fmt"

// OrderBy represents an ORDER BY clause
type OrderBy struct {
	Column    string
	Direction string
}

// ToSQL returns the SQL representation of the ORDER BY clause
func (o OrderBy) ToSQL() string {
	return fmt.Sprintf("%s %s", o.Column, o.Direction)
}
