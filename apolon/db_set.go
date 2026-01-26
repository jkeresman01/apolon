package apolon

import "github.com/jkeresman01/apolon/apolon-shared"

// DbSet provides a typed entry point for querying entities
type DbSet[T any] struct {
	db *DB
}

// newDbSet creates a new DbSet for the given type
func newDbSet[T any](db *DB) *DbSet[T] {
	return &DbSet[T]{db: db}
}

// Query returns a new query builder for this entity type
func (s *DbSet[T]) Query() *Query[T] {
	return newQuery[T](s.db)
}

// Where creates a new query with the given condition
func (s *DbSet[T]) Where(c shared.Condition) *Query[T] {
	return newQuery[T](s.db).Where(c)
}

// OrderBy creates a new query with the given ordering
func (s *DbSet[T]) OrderBy(o shared.OrderBy) *Query[T] {
	return newQuery[T](s.db).OrderBy(o)
}

// ToList returns all entities of this type
func (s *DbSet[T]) ToList() ([]T, error) {
	return newQuery[T](s.db).ToList()
}

// First returns the first entity or nil
func (s *DbSet[T]) First() (*T, error) {
	return newQuery[T](s.db).First()
}

// Count returns the total count of entities
func (s *DbSet[T]) Count() (int, error) {
	return newQuery[T](s.db).Count()
}
