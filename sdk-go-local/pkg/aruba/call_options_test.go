package aruba

import (
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"k8s.io/utils/ptr"
)

func TestCallOptions_Individual(t *testing.T) {
	opts := applyCallOptions([]CallOption{
		WithFilter("name eq foo"),
		WithSort("name asc"),
		WithProjection("id,name"),
		WithLimit(25),
		WithOffset(50),
		WithAPIVersion("2024-06-01"),
	})

	if opts.filter == nil || *opts.filter != "name eq foo" {
		t.Errorf("filter = %v", opts.filter)
	}
	if opts.sort == nil || *opts.sort != "name asc" {
		t.Errorf("sort = %v", opts.sort)
	}
	if opts.projection == nil || *opts.projection != "id,name" {
		t.Errorf("projection = %v", opts.projection)
	}
	if opts.limit == nil || *opts.limit != 25 {
		t.Errorf("limit = %v", opts.limit)
	}
	if opts.offset == nil || *opts.offset != 50 {
		t.Errorf("offset = %v", opts.offset)
	}
	if opts.apiVersion == nil || *opts.apiVersion != "2024-06-01" {
		t.Errorf("apiVersion = %v", opts.apiVersion)
	}
}

func TestCallOptions_ToRequestParameters(t *testing.T) {
	opts := applyCallOptions([]CallOption{
		WithFilter("x"),
		WithLimit(10),
	})
	rp := opts.toRequestParameters()
	if rp == nil {
		t.Fatal("toRequestParameters() returned nil")
	}
	if rp.Filter == nil || *rp.Filter != "x" {
		t.Errorf("Filter = %v", rp.Filter)
	}
	if rp.Limit == nil || *rp.Limit != 10 {
		t.Errorf("Limit = %v", rp.Limit)
	}
}

func TestCallOptions_Empty(t *testing.T) {
	opts := applyCallOptions(nil)
	rp := opts.toRequestParameters()
	if rp == nil {
		t.Fatal("toRequestParameters() returned nil for empty options")
	}
}

// WithRawParameters followed by WithFilter: the explicit option wins.
func TestCallOptions_RawThenExplicit(t *testing.T) {
	raw := &types.RequestParameters{
		Filter: ptr.To("from-raw"),
	}
	opts := applyCallOptions([]CallOption{
		WithRawParameters(raw),
		WithFilter("explicit"),
	})
	if opts.filter == nil || *opts.filter != "explicit" {
		t.Errorf("expected 'explicit', got %v", opts.filter)
	}
}

// WithFilter followed by WithRawParameters: the raw param wins.
func TestCallOptions_ExplicitThenRaw(t *testing.T) {
	raw := &types.RequestParameters{
		Filter: ptr.To("from-raw"),
	}
	opts := applyCallOptions([]CallOption{
		WithFilter("explicit"),
		WithRawParameters(raw),
	})
	if opts.filter == nil || *opts.filter != "from-raw" {
		t.Errorf("expected 'from-raw', got %v", opts.filter)
	}
}

// WithRawParameters with nil Limit does not clobber a previously set WithLimit.
func TestCallOptions_RawNilLimitPreservesExplicit(t *testing.T) {
	raw := &types.RequestParameters{
		Filter: ptr.To("from-raw"),
		// Limit is nil
	}
	opts := applyCallOptions([]CallOption{
		WithLimit(99),
		WithRawParameters(raw),
	})
	if opts.limit == nil || *opts.limit != 99 {
		t.Errorf("limit should still be 99, got %v", opts.limit)
	}
}

func TestCallOptions_WithRawParametersNil(t *testing.T) {
	opts := applyCallOptions([]CallOption{
		WithFilter("keep"),
		WithRawParameters(nil),
	})
	if opts.filter == nil || *opts.filter != "keep" {
		t.Errorf("expected filter 'keep' after WithRawParameters(nil), got %v", opts.filter)
	}
}
