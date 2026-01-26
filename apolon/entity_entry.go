package apolon

import (
	"reflect"

	"github.com/jkeresman01/apolon/apolon-shared"
)

// EntityEntry tracks a single entity and its state
type EntityEntry struct {
	Entity         any
	State          shared.EntityState
	OriginalValues map[string]any
	entityType     reflect.Type
	pkField        string
	pkValue        any
}

// newEntityEntry creates a new entity entry
func newEntityEntry(entity any, state shared.EntityState) *EntityEntry {
	entry := &EntityEntry{
		Entity:         entity,
		State:          state,
		OriginalValues: make(map[string]any),
	}
	entry.captureMetadata()
	if state == shared.Unchanged || state == shared.Modified {
		entry.captureOriginalValues()
	}
	return entry
}

// captureMetadata extracts type and primary key information
func (e *EntityEntry) captureMetadata() {
	v := reflect.ValueOf(e.Entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	e.entityType = v.Type()

	// Find primary key field (marked with ,pk in apolon tag)
	for i := 0; i < e.entityType.NumField(); i++ {
		field := e.entityType.Field(i)
		tag := field.Tag.Get("apolon")
		if containsPK(tag) {
			e.pkField = field.Name
			e.pkValue = v.Field(i).Interface()
			break
		}
	}

	// If no pk tag found, default to "ID" field
	if e.pkField == "" {
		if idField, ok := e.entityType.FieldByName("ID"); ok {
			e.pkField = idField.Name
			e.pkValue = v.FieldByName("ID").Interface()
		}
	}
}

// captureOriginalValues stores a snapshot of all field values
func (e *EntityEntry) captureOriginalValues() {
	v := reflect.ValueOf(e.Entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		fieldName := e.entityType.Field(i).Name
		e.OriginalValues[fieldName] = v.Field(i).Interface()
	}
}

// GetPrimaryKey returns the primary key value
func (e *EntityEntry) GetPrimaryKey() any {
	v := reflect.ValueOf(e.Entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if e.pkField != "" {
		return v.FieldByName(e.pkField).Interface()
	}
	return nil
}

// GetChangedProperties returns a map of properties that have changed
func (e *EntityEntry) GetChangedProperties() map[string]any {
	if e.State != shared.Modified && e.State != shared.Unchanged {
		return nil
	}

	changed := make(map[string]any)
	v := reflect.ValueOf(e.Entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		fieldName := e.entityType.Field(i).Name
		currentValue := v.Field(i).Interface()
		originalValue, exists := e.OriginalValues[fieldName]

		if exists && !reflect.DeepEqual(currentValue, originalValue) {
			changed[fieldName] = currentValue
		}
	}

	return changed
}

// HasChanges returns true if the entity has been modified
func (e *EntityEntry) HasChanges() bool {
	if e.State == shared.Added || e.State == shared.Deleted {
		return true
	}
	return len(e.GetChangedProperties()) > 0
}

// DetectChanges updates the state if entity has been modified
func (e *EntityEntry) DetectChanges() {
	if e.State == shared.Unchanged {
		if len(e.GetChangedProperties()) > 0 {
			e.State = shared.Modified
		}
	}
}

// AcceptChanges marks the entity as unchanged and updates original values
func (e *EntityEntry) AcceptChanges() {
	e.State = shared.Unchanged
	e.captureOriginalValues()
}

// containsPK checks if an apolon tag contains the pk option
func containsPK(tag string) bool {
	for _, part := range splitTag(tag) {
		if part == "pk" {
			return true
		}
	}
	return false
}

// splitTag splits an apolon tag by comma
func splitTag(tag string) []string {
	if tag == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i <= len(tag); i++ {
		if i == len(tag) || tag[i] == ',' {
			if start < i {
				parts = append(parts, tag[start:i])
			}
			start = i + 1
		}
	}
	return parts
}
