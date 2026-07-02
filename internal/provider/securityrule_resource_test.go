package provider

import (
	"testing"
)

func TestNormalizeProtocol(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"any", "Any"},
		{"ANY", "Any"},
		{"Any", "Any"},
		{"tcp", "TCP"},
		{"TCP", "TCP"},
		{"udp", "UDP"},
		{"UDP", "UDP"},
		{"icmp", "ICMP"},
		{"ICMP", "ICMP"},
		{"", ""},
		{"other", "Other"},
	}
	for _, tc := range cases {
		if got := normalizeProtocol(tc.in); got != tc.want {
			t.Errorf("normalizeProtocol(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeTargetKind(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"IP", "IP"},
		{"ip", "IP"},
		{"Ip", "IP"},
		{"SecurityGroup", "SecurityGroup"},
		{"securitygroup", "SecurityGroup"},
		{"SECURITYGROUP", "SecurityGroup"},
		{"", ""},
		{"unknown", "unknown"},
	}
	for _, tc := range cases {
		if got := normalizeTargetKind(tc.in); got != tc.want {
			t.Errorf("normalizeTargetKind(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
