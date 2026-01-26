package shared

import "time"

// BaseField contains common properties for all field types
type BaseField struct {
	Table  string
	Column string
}

// IntField for int, int32, int64 columns
type IntField struct {
	BaseField
}

// Eq returns a condition for column = value
func (f IntField) Eq(val int) Condition {
	return &SimpleCondition{f.Column, "=", val}
}

// Neq returns a condition for column != value
func (f IntField) Neq(val int) Condition {
	return &SimpleCondition{f.Column, "!=", val}
}

// Gt returns a condition for column > value
func (f IntField) Gt(val int) Condition {
	return &SimpleCondition{f.Column, ">", val}
}

// Gte returns a condition for column >= value
func (f IntField) Gte(val int) Condition {
	return &SimpleCondition{f.Column, ">=", val}
}

// Lt returns a condition for column < value
func (f IntField) Lt(val int) Condition {
	return &SimpleCondition{f.Column, "<", val}
}

// Lte returns a condition for column <= value
func (f IntField) Lte(val int) Condition {
	return &SimpleCondition{f.Column, "<=", val}
}

// In returns a condition for column IN (values...)
func (f IntField) In(vals ...int) Condition {
	anyVals := make([]any, len(vals))
	for i, v := range vals {
		anyVals[i] = v
	}
	return &InCondition{f.Column, anyVals}
}

// Between returns a condition for column BETWEEN low AND high
func (f IntField) Between(low, high int) Condition {
	return &BetweenCondition{f.Column, low, high}
}

// IsNull returns a condition for column IS NULL
func (f IntField) IsNull() Condition {
	return &NullCondition{f.Column, true}
}

// IsNotNull returns a condition for column IS NOT NULL
func (f IntField) IsNotNull() Condition {
	return &NullCondition{f.Column, false}
}

// Asc returns an ascending ORDER BY clause
func (f IntField) Asc() OrderBy {
	return OrderBy{f.Column, "ASC"}
}

// Desc returns a descending ORDER BY clause
func (f IntField) Desc() OrderBy {
	return OrderBy{f.Column, "DESC"}
}

// Int64Field for int64 columns
type Int64Field struct {
	BaseField
}

// Eq returns a condition for column = value
func (f Int64Field) Eq(val int64) Condition {
	return &SimpleCondition{f.Column, "=", val}
}

// Neq returns a condition for column != value
func (f Int64Field) Neq(val int64) Condition {
	return &SimpleCondition{f.Column, "!=", val}
}

// Gt returns a condition for column > value
func (f Int64Field) Gt(val int64) Condition {
	return &SimpleCondition{f.Column, ">", val}
}

// Gte returns a condition for column >= value
func (f Int64Field) Gte(val int64) Condition {
	return &SimpleCondition{f.Column, ">=", val}
}

// Lt returns a condition for column < value
func (f Int64Field) Lt(val int64) Condition {
	return &SimpleCondition{f.Column, "<", val}
}

// Lte returns a condition for column <= value
func (f Int64Field) Lte(val int64) Condition {
	return &SimpleCondition{f.Column, "<=", val}
}

// In returns a condition for column IN (values...)
func (f Int64Field) In(vals ...int64) Condition {
	anyVals := make([]any, len(vals))
	for i, v := range vals {
		anyVals[i] = v
	}
	return &InCondition{f.Column, anyVals}
}

// Between returns a condition for column BETWEEN low AND high
func (f Int64Field) Between(low, high int64) Condition {
	return &BetweenCondition{f.Column, low, high}
}

// IsNull returns a condition for column IS NULL
func (f Int64Field) IsNull() Condition {
	return &NullCondition{f.Column, true}
}

// IsNotNull returns a condition for column IS NOT NULL
func (f Int64Field) IsNotNull() Condition {
	return &NullCondition{f.Column, false}
}

// Asc returns an ascending ORDER BY clause
func (f Int64Field) Asc() OrderBy {
	return OrderBy{f.Column, "ASC"}
}

// Desc returns a descending ORDER BY clause
func (f Int64Field) Desc() OrderBy {
	return OrderBy{f.Column, "DESC"}
}

// StringField for string columns
type StringField struct {
	BaseField
}

// Eq returns a condition for column = value
func (f StringField) Eq(val string) Condition {
	return &SimpleCondition{f.Column, "=", val}
}

// Neq returns a condition for column != value
func (f StringField) Neq(val string) Condition {
	return &SimpleCondition{f.Column, "!=", val}
}

// Contains returns a condition for column LIKE %value%
func (f StringField) Contains(val string) Condition {
	return &LikeCondition{f.Column, "%" + val + "%"}
}

// StartsWith returns a condition for column LIKE value%
func (f StringField) StartsWith(val string) Condition {
	return &LikeCondition{f.Column, val + "%"}
}

// EndsWith returns a condition for column LIKE %value
func (f StringField) EndsWith(val string) Condition {
	return &LikeCondition{f.Column, "%" + val}
}

// Like returns a condition for column LIKE pattern
func (f StringField) Like(pattern string) Condition {
	return &LikeCondition{f.Column, pattern}
}

// In returns a condition for column IN (values...)
func (f StringField) In(vals ...string) Condition {
	anyVals := make([]any, len(vals))
	for i, v := range vals {
		anyVals[i] = v
	}
	return &InCondition{f.Column, anyVals}
}

// IsNull returns a condition for column IS NULL
func (f StringField) IsNull() Condition {
	return &NullCondition{f.Column, true}
}

// IsNotNull returns a condition for column IS NOT NULL
func (f StringField) IsNotNull() Condition {
	return &NullCondition{f.Column, false}
}

// Asc returns an ascending ORDER BY clause
func (f StringField) Asc() OrderBy {
	return OrderBy{f.Column, "ASC"}
}

// Desc returns a descending ORDER BY clause
func (f StringField) Desc() OrderBy {
	return OrderBy{f.Column, "DESC"}
}

// BoolField for boolean columns
type BoolField struct {
	BaseField
}

// Eq returns a condition for column = value
func (f BoolField) Eq(val bool) Condition {
	return &SimpleCondition{f.Column, "=", val}
}

// IsTrue returns a condition for column = true
func (f BoolField) IsTrue() Condition {
	return &SimpleCondition{f.Column, "=", true}
}

// IsFalse returns a condition for column = false
func (f BoolField) IsFalse() Condition {
	return &SimpleCondition{f.Column, "=", false}
}

// IsNull returns a condition for column IS NULL
func (f BoolField) IsNull() Condition {
	return &NullCondition{f.Column, true}
}

// IsNotNull returns a condition for column IS NOT NULL
func (f BoolField) IsNotNull() Condition {
	return &NullCondition{f.Column, false}
}

// Asc returns an ascending ORDER BY clause
func (f BoolField) Asc() OrderBy {
	return OrderBy{f.Column, "ASC"}
}

// Desc returns a descending ORDER BY clause
func (f BoolField) Desc() OrderBy {
	return OrderBy{f.Column, "DESC"}
}

// FloatField for float32, float64 columns
type FloatField struct {
	BaseField
}

// Eq returns a condition for column = value
func (f FloatField) Eq(val float64) Condition {
	return &SimpleCondition{f.Column, "=", val}
}

// Neq returns a condition for column != value
func (f FloatField) Neq(val float64) Condition {
	return &SimpleCondition{f.Column, "!=", val}
}

// Gt returns a condition for column > value
func (f FloatField) Gt(val float64) Condition {
	return &SimpleCondition{f.Column, ">", val}
}

// Gte returns a condition for column >= value
func (f FloatField) Gte(val float64) Condition {
	return &SimpleCondition{f.Column, ">=", val}
}

// Lt returns a condition for column < value
func (f FloatField) Lt(val float64) Condition {
	return &SimpleCondition{f.Column, "<", val}
}

// Lte returns a condition for column <= value
func (f FloatField) Lte(val float64) Condition {
	return &SimpleCondition{f.Column, "<=", val}
}

// Between returns a condition for column BETWEEN low AND high
func (f FloatField) Between(low, high float64) Condition {
	return &BetweenCondition{f.Column, low, high}
}

// IsNull returns a condition for column IS NULL
func (f FloatField) IsNull() Condition {
	return &NullCondition{f.Column, true}
}

// IsNotNull returns a condition for column IS NOT NULL
func (f FloatField) IsNotNull() Condition {
	return &NullCondition{f.Column, false}
}

// Asc returns an ascending ORDER BY clause
func (f FloatField) Asc() OrderBy {
	return OrderBy{f.Column, "ASC"}
}

// Desc returns a descending ORDER BY clause
func (f FloatField) Desc() OrderBy {
	return OrderBy{f.Column, "DESC"}
}

// TimeField for time.Time columns
type TimeField struct {
	BaseField
}

// Eq returns a condition for column = value
func (f TimeField) Eq(val time.Time) Condition {
	return &SimpleCondition{f.Column, "=", val}
}

// Neq returns a condition for column != value
func (f TimeField) Neq(val time.Time) Condition {
	return &SimpleCondition{f.Column, "!=", val}
}

// Before returns a condition for column < value
func (f TimeField) Before(val time.Time) Condition {
	return &SimpleCondition{f.Column, "<", val}
}

// After returns a condition for column > value
func (f TimeField) After(val time.Time) Condition {
	return &SimpleCondition{f.Column, ">", val}
}

// BeforeOrEqual returns a condition for column <= value
func (f TimeField) BeforeOrEqual(val time.Time) Condition {
	return &SimpleCondition{f.Column, "<=", val}
}

// AfterOrEqual returns a condition for column >= value
func (f TimeField) AfterOrEqual(val time.Time) Condition {
	return &SimpleCondition{f.Column, ">=", val}
}

// Between returns a condition for column BETWEEN start AND end
func (f TimeField) Between(start, end time.Time) Condition {
	return &BetweenCondition{f.Column, start, end}
}

// IsNull returns a condition for column IS NULL
func (f TimeField) IsNull() Condition {
	return &NullCondition{f.Column, true}
}

// IsNotNull returns a condition for column IS NOT NULL
func (f TimeField) IsNotNull() Condition {
	return &NullCondition{f.Column, false}
}

// Asc returns an ascending ORDER BY clause
func (f TimeField) Asc() OrderBy {
	return OrderBy{f.Column, "ASC"}
}

// Desc returns a descending ORDER BY clause
func (f TimeField) Desc() OrderBy {
	return OrderBy{f.Column, "DESC"}
}
