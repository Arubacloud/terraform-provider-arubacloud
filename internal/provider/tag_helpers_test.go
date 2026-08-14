package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBillingPeriodFromAPI(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"hourly", "Hour"},
		{"monthly", "Month"},
		{"yearly", "Year"},
		{"Hour", "Hour"},       // already canonical
		{"Month", "Month"},     // already canonical
		{"Year", "Year"},       // already canonical
		{"unknown", "unknown"}, // passthrough
		{"", ""},               // empty passthrough
	}
	for _, tc := range cases {
		if got := billingPeriodFromAPI(tc.in); got != tc.want {
			t.Errorf("billingPeriodFromAPI(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTagsToList(t *testing.T) {
	cases := []struct {
		name     string
		input    []string
		wantLen  int
		wantTags []string
	}{
		{
			name:     "nil slice produces empty list",
			input:    nil,
			wantLen:  0,
			wantTags: []string{},
		},
		{
			name:     "empty slice produces empty list",
			input:    []string{},
			wantLen:  0,
			wantTags: []string{},
		},
		{
			name:     "single tag",
			input:    []string{"env:prod"},
			wantLen:  1,
			wantTags: []string{"env:prod"},
		},
		{
			name:     "multiple tags",
			input:    []string{"env:prod", "team:platform", "cost-center:123"},
			wantLen:  3,
			wantTags: []string{"env:prod", "team:platform", "cost-center:123"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TagsToList(tc.input)
			if got.IsNull() {
				t.Fatal("expected non-null list, got null")
			}
			if got.IsUnknown() {
				t.Fatal("expected non-unknown list, got unknown")
			}
			if got.ElementType(t.Context()) != types.StringType {
				t.Fatalf("expected element type StringType, got %T", got.ElementType(t.Context()))
			}
			if len(got.Elements()) != tc.wantLen {
				t.Fatalf("expected %d elements, got %d", tc.wantLen, len(got.Elements()))
			}
			var tags []string
			got.ElementsAs(t.Context(), &tags, false)
			for i, want := range tc.wantTags {
				if tags[i] != want {
					t.Errorf("element[%d]: got %q, want %q", i, tags[i], want)
				}
			}
		})
	}
}

func TestTagsToListPreserveNull(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		prior    types.List
		wantNull bool
		wantLen  int
		wantTags []string
	}{
		{
			name:     "API empty + prior null → null (no phantom diff for omitted attribute)",
			tags:     nil,
			prior:    types.ListNull(types.StringType),
			wantNull: true,
		},
		{
			name:     "API empty slice + prior null → null",
			tags:     []string{},
			prior:    types.ListNull(types.StringType),
			wantNull: true,
		},
		{
			name:     "API empty + prior empty list → empty list (user set tags = [])",
			tags:     nil,
			prior:    types.ListValueMust(types.StringType, []attr.Value{}),
			wantNull: false,
			wantLen:  0,
		},
		{
			name:     "API empty + prior had tags → empty list (tags cleared)",
			tags:     nil,
			prior:    TagsToList([]string{"old-tag"}),
			wantNull: false,
			wantLen:  0,
		},
		{
			name:     "API has tags + prior null → list with tags",
			tags:     []string{"new-tag"},
			prior:    types.ListNull(types.StringType),
			wantNull: false,
			wantLen:  1,
			wantTags: []string{"new-tag"},
		},
		{
			name:     "API has tags + prior matches → list with tags",
			tags:     []string{"tag-a", "tag-b"},
			prior:    TagsToList([]string{"tag-a", "tag-b"}),
			wantNull: false,
			wantLen:  2,
			wantTags: []string{"tag-a", "tag-b"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TagsToListPreserveNull(tc.tags, tc.prior)
			if tc.wantNull {
				if !got.IsNull() {
					t.Fatalf("expected null list, got %v", got)
				}
				return
			}
			if got.IsNull() {
				t.Fatal("expected non-null list, got null")
			}
			if len(got.Elements()) != tc.wantLen {
				t.Fatalf("expected %d elements, got %d", tc.wantLen, len(got.Elements()))
			}
			var tags []string
			got.ElementsAs(t.Context(), &tags, false)
			for i, want := range tc.wantTags {
				if tags[i] != want {
					t.Errorf("element[%d]: got %q, want %q", i, tags[i], want)
				}
			}
		})
	}
}

func TestListToTags(t *testing.T) {
	ctx := t.Context()

	makeList := func(tags ...string) types.List {
		vals := make([]attr.Value, len(tags))
		for i, tag := range tags {
			vals[i] = types.StringValue(tag)
		}
		return types.ListValueMust(types.StringType, vals)
	}

	cases := []struct {
		name     string
		list     types.List
		wantTags []string
		wantNil  bool
	}{
		{
			name:    "null list returns nil",
			list:    types.ListNull(types.StringType),
			wantNil: true,
		},
		{
			name:    "unknown list returns nil",
			list:    types.ListUnknown(types.StringType),
			wantNil: true,
		},
		{
			name:     "empty list returns empty slice",
			list:     makeList(),
			wantTags: []string{},
		},
		{
			name:     "single element",
			list:     makeList("env:prod"),
			wantTags: []string{"env:prod"},
		},
		{
			name:     "multiple elements",
			list:     makeList("env:prod", "team:platform"),
			wantTags: []string{"env:prod", "team:platform"},
		},
		{
			name:     "round-trip through TagsToList",
			list:     TagsToList([]string{"a", "b", "c"}),
			wantTags: []string{"a", "b", "c"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var d diag.Diagnostics
			got := ListToTags(ctx, tc.list, &d)
			if d.HasError() {
				t.Fatalf("unexpected diagnostics errors: %v", d)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tc.wantTags) {
				t.Fatalf("expected %d tags, got %d: %v", len(tc.wantTags), len(got), got)
			}
			for i, want := range tc.wantTags {
				if got[i] != want {
					t.Errorf("tag[%d]: got %q, want %q", i, got[i], want)
				}
			}
		})
	}
}
