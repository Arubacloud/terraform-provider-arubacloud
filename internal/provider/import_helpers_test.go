package provider

import (
	"testing"
)

func TestParseImportID_StrictSegmentMatching(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		format    string
		example   string
		n         int
		wantParts []string
		wantErr   bool
	}{
		{
			name:      "valid 2-part ID",
			id:        "proj-123/res-456",
			format:    "project_id/resource_id",
			example:   "p123/r456",
			n:         2,
			wantParts: []string{"proj-123", "res-456"},
			wantErr:   false,
		},
		{
			name:      "valid 3-part ID",
			id:        "proj-123/vpc-456/sub-789",
			format:    "project_id/vpc_id/subnet_id",
			example:   "p123/v456/s789",
			n:         3,
			wantParts: []string{"proj-123", "vpc-456", "sub-789"},
			wantErr:   false,
		},
		{
			name:    "extra segment rejected (upper-bound test)",
			id:      "proj-123/vpc-456/sub-789/extra-segment",
			format:  "project_id/vpc_id/subnet_id",
			example: "p123/v456/s789",
			n:       3,
			wantErr: true,
		},
		{
			name:    "too few segments rejected",
			id:      "proj-123",
			format:  "project_id/resource_id",
			example: "p123/r456",
			n:       2,
			wantErr: true,
		},
		{
			name:    "empty segment rejected",
			id:      "proj-123//sub-789",
			format:  "project_id/vpc_id/subnet_id",
			example: "p123/v456/s789",
			n:       3,
			wantErr: true,
		},
		{
			name:    "empty id rejected",
			id:      "",
			format:  "project_id/resource_id",
			example: "p123/r456",
			n:       2,
			wantErr: true,
		},
		{
			name:      "whitespace trimmed around id",
			id:        "  proj-123/vpc-456  ",
			format:    "project_id/vpc_id",
			example:   "p123/v456",
			n:         2,
			wantParts: []string{"proj-123", "vpc-456"},
			wantErr:   false,
		},
		{
			name:    "leading slash rejected",
			id:      "/proj-123/vpc-456",
			format:  "project_id/vpc_id",
			example: "p123/v456",
			n:       2,
			wantErr: true,
		},
		{
			name:    "trailing slash rejected",
			id:      "proj-123/vpc-456/",
			format:  "project_id/vpc_id",
			example: "p123/v456",
			n:       2,
			wantErr: true,
		},
		{
			name:      "segment whitespace trimmed",
			id:        "proj-123 / vpc-456",
			format:    "project_id/vpc_id",
			example:   "p123/v456",
			n:         2,
			wantParts: []string{"proj-123", "vpc-456"},
			wantErr:   false,
		},
		{
			name:    "invalid n=0 rejected",
			id:      "proj-123/vpc-456",
			format:  "project_id/vpc_id",
			example: "p123/v456",
			n:       0,
			wantErr: true,
		},
		{
			name:    "invalid n=-1 rejected",
			id:      "proj-123/vpc-456",
			format:  "project_id/vpc_id",
			example: "p123/v456",
			n:       -1,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseImportID(tc.id, tc.format, tc.example, tc.n)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for id %q with n=%d, got nil", tc.id, tc.n)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error for id %q, got: %v", tc.id, err)
			}
			if len(got) != len(tc.wantParts) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tc.wantParts))
			}
			for i, p := range got {
				if p != tc.wantParts[i] {
					t.Errorf("part[%d] = %q, want %q", i, p, tc.wantParts[i])
				}
			}
		})
	}
}
