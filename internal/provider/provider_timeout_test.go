package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseTimeout(t *testing.T) {
	defaultDur := 10 * time.Minute

	cases := []struct {
		name    string
		input   types.String
		wantDur time.Duration
	}{
		{"null returns default", types.StringNull(), defaultDur},
		{"unknown returns default", types.StringUnknown(), defaultDur},
		{"empty string returns default", types.StringValue(""), defaultDur},
		{"valid 5m", types.StringValue("5m"), 5 * time.Minute},
		{"valid 1h", types.StringValue("1h"), time.Hour},
		{"valid 30s", types.StringValue("30s"), 30 * time.Second},
		// invalid string: returns default (warning added to caller's diags, but
		// diag.Diagnostics is a slice passed by value so the caller sees nothing)
		{"invalid string returns default", types.StringValue("notaduration"), defaultDur},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var d diag.Diagnostics
			got := parseTimeout(tc.input, defaultDur, d)
			if got != tc.wantDur {
				t.Errorf("parseTimeout(%q) = %v, want %v", tc.input.ValueString(), got, tc.wantDur)
			}
		})
	}
}
