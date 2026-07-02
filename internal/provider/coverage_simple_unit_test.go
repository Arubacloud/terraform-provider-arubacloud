package provider

import (
	"testing"
)

// TestPtrToString verifies the ptrToString helper (defined in kaas_data_source.go)
// with both a non-nil and a nil pointer, covering the two branches of the function.
func TestPtrToString(t *testing.T) {
	s := "hello"
	if got := ptrToString(&s); got != "hello" {
		t.Errorf("ptrToString(&s): got %q, want %q", got, "hello")
	}
	if got := ptrToString(nil); got != "" {
		t.Errorf("ptrToString(nil): got %q, want %q", got, "")
	}
}
