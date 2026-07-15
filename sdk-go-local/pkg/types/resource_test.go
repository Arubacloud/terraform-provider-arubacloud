package types_test

import (
	"errors"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestResourceMetadataResponse_Validate(t *testing.T) {
	id := "res-123"
	uri := "/projects/p/providers/Aruba.Compute/cloudServers/res-123"
	name := "my-server"

	t.Run("all fields present", func(t *testing.T) {
		m := &types.ResourceMetadataResponse{
			ID:   &id,
			URI:  &uri,
			Name: &name,
		}
		if err := m.Validate(); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var m *types.ResourceMetadataResponse
		err := m.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected MetadataValidationError, got %T", err)
		}
	})

	t.Run("missing id nil", func(t *testing.T) {
		m := &types.ResourceMetadataResponse{URI: &uri, Name: &name}
		err := m.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected MetadataValidationError, got %T", err)
		}
		if len(mvErr.Missing) != 1 || mvErr.Missing[0] != "id" {
			t.Errorf("expected missing=[id], got %v", mvErr.Missing)
		}
	})

	t.Run("missing id empty string", func(t *testing.T) {
		empty := ""
		m := &types.ResourceMetadataResponse{ID: &empty, URI: &uri, Name: &name}
		err := m.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected MetadataValidationError, got %T", err)
		}
		if len(mvErr.Missing) != 1 || mvErr.Missing[0] != "id" {
			t.Errorf("expected missing=[id], got %v", mvErr.Missing)
		}
	})

	t.Run("missing name nil", func(t *testing.T) {
		m := &types.ResourceMetadataResponse{ID: &id, URI: &uri}
		err := m.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected MetadataValidationError, got %T", err)
		}
		if len(mvErr.Missing) != 1 || mvErr.Missing[0] != "name" {
			t.Errorf("expected missing=[name], got %v", mvErr.Missing)
		}
	})

	t.Run("missing name empty string", func(t *testing.T) {
		empty := ""
		m := &types.ResourceMetadataResponse{ID: &id, URI: &uri, Name: &empty}
		err := m.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected MetadataValidationError, got %T", err)
		}
		if len(mvErr.Missing) != 1 || mvErr.Missing[0] != "name" {
			t.Errorf("expected missing=[name], got %v", mvErr.Missing)
		}
	})

	t.Run("all fields missing", func(t *testing.T) {
		m := &types.ResourceMetadataResponse{}
		err := m.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected MetadataValidationError, got %T", err)
		}
		if len(mvErr.Missing) != 2 {
			t.Errorf("expected 2 missing fields, got %v", mvErr.Missing)
		}
	})

	t.Run("error message lists missing fields", func(t *testing.T) {
		m := &types.ResourceMetadataResponse{Name: &name}
		err := m.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		msg := err.Error()
		if msg == "" {
			t.Error("expected non-empty error message")
		}
	})
}
