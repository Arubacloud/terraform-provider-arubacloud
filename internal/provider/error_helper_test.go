package provider

import "testing"

func TestBillingPeriodToAPI(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Hour", "hourly"},
		{"Month", "monthly"},
		{"Year", "yearly"},
		{"hourly", "hourly"},   // already in API form
		{"monthly", "monthly"}, // already in API form
		{"yearly", "yearly"},   // already in API form
		{"unknown", "unknown"}, // passthrough
		{"", ""},               // empty passthrough
	}
	for _, tc := range cases {
		if got := billingPeriodToAPI(tc.in); got != tc.want {
			t.Errorf("billingPeriodToAPI(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBillingPeriodFromAPI(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"hourly", "Hour"},
		{"monthly", "Month"},
		{"yearly", "Year"},
		{"Hour", "Hour"},     // already canonical
		{"Month", "Month"},   // already canonical
		{"Year", "Year"},     // already canonical
		{"unknown", "unknown"}, // passthrough
		{"", ""},               // empty passthrough
	}
	for _, tc := range cases {
		if got := billingPeriodFromAPI(tc.in); got != tc.want {
			t.Errorf("billingPeriodFromAPI(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
