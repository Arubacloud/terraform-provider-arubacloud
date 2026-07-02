package provider

import (
	"net"
	"testing"
)

// ── cidrContains ──────────────────────────────────────────────────────────────

func TestCidrContains(t *testing.T) {
	parseCIDR := func(s string) *net.IPNet {
		t.Helper()
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatalf("ParseCIDR(%q): %v", s, err)
		}
		return n
	}

	cases := []struct {
		parent string
		child  string
		want   bool
	}{
		// Subnet is contained in the parent.
		{"10.0.0.0/8", "10.1.2.0/24", true},
		// Equal networks are considered contained.
		{"10.0.0.0/24", "10.0.0.0/24", true},
		// Child prefix longer than parent — still a subset.
		{"192.168.0.0/16", "192.168.1.0/24", true},
		// Child has a different network address — not contained.
		{"10.0.0.0/24", "10.0.1.0/24", false},
		// Parent is a subnet of child (narrower mask) — not contained.
		{"10.0.0.0/24", "10.0.0.0/16", false},
		// IPv4 vs IPv6 — different bit widths, never contained.
		{"10.0.0.0/8", "::1/128", false},
	}

	for _, tc := range cases {
		parent, child := parseCIDR(tc.parent), parseCIDR(tc.child)
		if got := cidrContains(parent, child); got != tc.want {
			t.Errorf("cidrContains(%q, %q) = %v, want %v", tc.parent, tc.child, got, tc.want)
		}
	}
}
