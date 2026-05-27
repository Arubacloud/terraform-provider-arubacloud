package provider

import "testing"

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
