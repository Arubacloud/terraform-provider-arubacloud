package types_test

import (
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestState_IsTransitory(t *testing.T) {
	transitory := []types.State{
		types.StateInCreation,
		types.StateCreating,
		types.StateUpdating,
		types.StateProvisioning,
		types.StateDeleting,
		types.StateDisabling,
		types.StateEnabling,
	}
	for _, s := range transitory {
		if !s.IsTransitory() {
			t.Errorf("IsTransitory(%q) = false, want true", s)
		}
		if s.IsFailure() {
			t.Errorf("IsFailure(%q) = true, want false", s)
		}
		if s.IsBound() {
			t.Errorf("IsBound(%q) = true, want false", s)
		}
		if s.IsAvailable() {
			t.Errorf("IsAvailable(%q) = true, want false", s)
		}
	}

	settled := []types.State{
		types.StateActive, types.StateRunning, types.StateStopped,
		types.StateNotUsed, types.StateReserved, types.StateInUse, types.StateUsed,
		types.StateDeleted, types.StateFailed, types.StateError, types.StateDisabled,
	}
	for _, s := range settled {
		if s.IsTransitory() {
			t.Errorf("IsTransitory(%q) = true, want false", s)
		}
	}
}

func TestState_IsFailure(t *testing.T) {
	failures := []types.State{types.StateFailed, types.StateError, types.StateDisabled}
	for _, s := range failures {
		if !s.IsFailure() {
			t.Errorf("IsFailure(%q) = false, want true", s)
		}
		if s.IsTransitory() {
			t.Errorf("IsTransitory(%q) = true, want false", s)
		}
		if s.IsBound() {
			t.Errorf("IsBound(%q) = true, want false", s)
		}
		if s.IsAvailable() {
			t.Errorf("IsAvailable(%q) = true, want false", s)
		}
	}

	nonFailures := []types.State{
		types.StateActive, types.StateRunning, types.StateStopped,
		types.StateNotUsed, types.StateReserved, types.StateInUse, types.StateUsed,
	}
	for _, s := range nonFailures {
		if s.IsFailure() {
			t.Errorf("IsFailure(%q) = true, want false", s)
		}
	}
}

func TestState_IsBound(t *testing.T) {
	bound := []types.State{types.StateReserved, types.StateInUse, types.StateUsed}
	for _, s := range bound {
		if !s.IsBound() {
			t.Errorf("IsBound(%q) = false, want true", s)
		}
	}

	free := []types.State{
		types.StateActive, types.StateRunning, types.StateStopped,
		types.StateNotUsed, types.StateDeleted,
		types.StateFailed, types.StateError, types.StateDisabled,
	}
	for _, s := range free {
		if s.IsBound() {
			t.Errorf("IsBound(%q) = true, want false", s)
		}
	}
}

func TestState_IsAvailable(t *testing.T) {
	if !types.StateNotUsed.IsAvailable() {
		t.Error("IsAvailable(NotUsed) = false, want true")
	}

	notAvailable := []types.State{
		types.StateActive, types.StateRunning, types.StateReserved,
		types.StateInUse, types.StateUsed, types.StateStopped,
		types.StateDeleted, types.StateFailed,
	}
	for _, s := range notAvailable {
		if s.IsAvailable() {
			t.Errorf("IsAvailable(%q) = true, want false", s)
		}
	}
}

func TestState_IsOperational(t *testing.T) {
	operational := []types.State{
		types.StateActive, types.StateRunning, types.StateInUse, types.StateUsed,
	}
	for _, s := range operational {
		if !s.IsOperational() {
			t.Errorf("IsOperational(%q) = false, want true", s)
		}
	}

	nonOperational := []types.State{
		types.StateStopped, types.StateNotUsed, types.StateReserved,
		types.StateDeleted, types.StateFailed, types.StateError, types.StateDisabled,
	}
	for _, s := range nonOperational {
		if s.IsOperational() {
			t.Errorf("IsOperational(%q) = true, want false", s)
		}
	}
}

func TestState_Reserved_Membership(t *testing.T) {
	s := types.StateReserved
	if s.IsTransitory() {
		t.Error("Reserved.IsTransitory() = true, want false")
	}
	if s.IsFailure() {
		t.Error("Reserved.IsFailure() = true, want false")
	}
	if !s.IsBound() {
		t.Error("Reserved.IsBound() = false, want true")
	}
	if s.IsAvailable() {
		t.Error("Reserved.IsAvailable() = true, want false")
	}
	if s.IsOperational() {
		t.Error("Reserved.IsOperational() = true, want false")
	}
}

func TestState_UnknownAndEmpty(t *testing.T) {
	for _, s := range []types.State{"", types.State("Bogus"), types.State("unknown")} {
		if s.IsTransitory() {
			t.Errorf("IsTransitory(%q) = true for unknown state", s)
		}
		if s.IsFailure() {
			t.Errorf("IsFailure(%q) = true for unknown state", s)
		}
		if s.IsBound() {
			t.Errorf("IsBound(%q) = true for unknown state", s)
		}
		if s.IsAvailable() {
			t.Errorf("IsAvailable(%q) = true for unknown state", s)
		}
		if s.IsOperational() {
			t.Errorf("IsOperational(%q) = true for unknown state", s)
		}
	}
}

func TestState_DeletedAndStopped_NoSet(t *testing.T) {
	for _, s := range []types.State{types.StateDeleted, types.StateStopped} {
		if s.IsTransitory() || s.IsFailure() || s.IsBound() || s.IsAvailable() || s.IsOperational() {
			t.Errorf("State %q should belong to no predicate set", s)
		}
	}
}
