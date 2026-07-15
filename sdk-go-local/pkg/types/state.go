package types

// State is the lifecycle state of an Aruba Cloud resource (ResourceStatusResponse.State).
type State string

const (
	// Transitory — an operation is in progress.
	StateInCreation   State = "InCreation"
	StateCreating     State = "Creating"
	StateUpdating     State = "Updating"
	StateProvisioning State = "Provisioning"
	StateDeleting     State = "Deleting"
	StateDisabling    State = "Disabling"
	StateEnabling     State = "Enabling"

	// Settled — no operation in progress.
	StateActive   State = "Active"
	StateRunning  State = "Running"
	StateStopped  State = "Stopped"
	StateNotUsed  State = "NotUsed"  // free to be bound
	StateReserved State = "Reserved" // bound as a dependency, not actively in use
	StateInUse    State = "InUse"
	StateUsed     State = "Used"
	StateDeleted  State = "Deleted"

	// Settled + failure.
	StateFailed   State = "Failed"
	StateError    State = "Error"
	StateDisabled State = "Disabled" // requires manual re-enablement
)

// IsTransitory reports whether an operation is in progress.
func (s State) IsTransitory() bool {
	switch s {
	case StateInCreation, StateCreating, StateUpdating, StateProvisioning,
		StateDeleting, StateDisabling, StateEnabling:
		return true
	}
	return false
}

// IsFailure reports whether the resource is in a failure state.
// Failed and Error indicate a provisioning/operational fault; Disabled
// indicates the server has administratively disabled the resource.
func (s State) IsFailure() bool {
	switch s {
	case StateFailed, StateError, StateDisabled:
		return true
	}
	return false
}

// IsBound reports whether the resource is claimed (not free to be reused).
// Reserved means bound as a dependency but not actively used; InUse and Used
// mean actively attached or consumed.
func (s State) IsBound() bool {
	switch s {
	case StateReserved, StateInUse, StateUsed:
		return true
	}
	return false
}

// IsAvailable reports whether the resource is free to be bound.
func (s State) IsAvailable() bool {
	return s == StateNotUsed
}

// IsOperational reports whether the resource is actively serving traffic or
// compute (Active, Running) or is attached and in active use (InUse, Used).
func (s State) IsOperational() bool {
	switch s {
	case StateActive, StateRunning, StateInUse, StateUsed:
		return true
	}
	return false
}
