package apolon

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/jkeresman01/apolon/apolon-shared"
)

// ChangeTracker tracks all entity changes for a DbContext
type ChangeTracker struct {
	entries map[string]*EntityEntry // key: "TypeName:PKValue"
	mu      sync.RWMutex
}

func newChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		entries: make(map[string]*EntityEntry),
	}
}

// Track begins tracking an entity with the given state
func (ct *ChangeTracker) Track(entity any, state shared.EntityState) *EntityEntry {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	entry := newEntityEntry(entity, state)
	key := ct.makeKey(entity, entry.GetPrimaryKey())
	ct.entries[key] = entry
	return entry
}

// TrackRange tracks multiple entities with the given state
func (ct *ChangeTracker) TrackRange(entities any, state shared.EntityState) {
	v := reflect.ValueOf(entities)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				ct.Track(elem.Interface(), state)
			} else {
				ct.Track(elem.Addr().Interface(), state)
			}
		}
	}
}

// GetEntry returns the entry for an entity if it's being tracked
func (ct *ChangeTracker) GetEntry(entity any) *EntityEntry {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// First try to find by pointer identity
	for _, entry := range ct.entries {
		if entry.Entity == entity {
			return entry
		}
	}

	// Then try by type and PK
	tempEntry := newEntityEntry(entity, shared.Detached)
	key := ct.makeKey(entity, tempEntry.GetPrimaryKey())
	return ct.entries[key]
}

// GetEntryByKey returns the entry for an entity by type and primary key
func (ct *ChangeTracker) GetEntryByKey(entityType reflect.Type, pk any) *EntityEntry {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	key := fmt.Sprintf("%s:%v", entityType.Name(), pk)
	return ct.entries[key]
}

// Entries returns all tracked entries
func (ct *ChangeTracker) Entries() []*EntityEntry {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make([]*EntityEntry, 0, len(ct.entries))
	for _, entry := range ct.entries {
		result = append(result, entry)
	}
	return result
}

// EntriesByState returns all entries with the given state
func (ct *ChangeTracker) EntriesByState(state shared.EntityState) []*EntityEntry {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	var result []*EntityEntry
	for _, entry := range ct.entries {
		if entry.State == state {
			result = append(result, entry)
		}
	}
	return result
}

// HasChanges returns true if there are any pending changes
func (ct *ChangeTracker) HasChanges() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	for _, entry := range ct.entries {
		if entry.State == shared.Added || entry.State == shared.Deleted {
			return true
		}
		if entry.State == shared.Modified || entry.State == shared.Unchanged {
			if entry.HasChanges() {
				return true
			}
		}
	}
	return false
}

// DetectChanges scans all tracked entities for changes
func (ct *ChangeTracker) DetectChanges() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	for _, entry := range ct.entries {
		entry.DetectChanges()
	}
}

// AcceptAllChanges marks all entities as unchanged
func (ct *ChangeTracker) AcceptAllChanges() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	for key, entry := range ct.entries {
		if entry.State == shared.Deleted {
			delete(ct.entries, key)
		} else {
			entry.AcceptChanges()
		}
	}
}

// Clear removes all tracked entities
func (ct *ChangeTracker) Clear() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.entries = make(map[string]*EntityEntry)
}

// Untrack stops tracking an entity
func (ct *ChangeTracker) Untrack(entity any) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	tempEntry := newEntityEntry(entity, shared.Detached)
	key := ct.makeKey(entity, tempEntry.GetPrimaryKey())
	delete(ct.entries, key)
}

// makeKey creates a unique key for an entity based on type and PK
// For new entities (zero PK), uses pointer address to avoid collisions
func (ct *ChangeTracker) makeKey(entity any, pk any) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if PK is zero value - if so, use pointer address
	if isZeroValue(pk) {
		return fmt.Sprintf("%s:ptr:%p", t.Name(), entity)
	}
	return fmt.Sprintf("%s:%v", t.Name(), pk)
}

// GetOrTrack returns an existing entry or creates a new one
func (ct *ChangeTracker) GetOrTrack(entity any, state shared.EntityState) *EntityEntry {
	if entry := ct.GetEntry(entity); entry != nil {
		return entry
	}
	return ct.Track(entity, state)
}
