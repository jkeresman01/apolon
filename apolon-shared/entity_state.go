package shared

// EntityState represents the state of an entity in the change tracker
type EntityState int

const (
	// Detached - Entity is not being tracked
	Detached EntityState = iota
	// Unchanged - Entity is tracked and has not been modified since loading
	Unchanged
	// Added - Entity is tracked and will be inserted on SaveChanges
	Added
	// Modified - Entity is tracked and has been modified since loading
	Modified
	// Deleted - Entity is tracked and will be deleted on SaveChanges
	Deleted
)

func (s EntityState) String() string {
	switch s {
	case Detached:
		return "Detached"
	case Unchanged:
		return "Unchanged"
	case Added:
		return "Added"
	case Modified:
		return "Modified"
	case Deleted:
		return "Deleted"
	default:
		return "Unknown"
	}
}
