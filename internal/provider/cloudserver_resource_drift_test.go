package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveAPIStringRef(t *testing.T) {
	stateVal := types.StringValue("/projects/p/vpcs/v")

	tests := []struct {
		name     string
		apiValue string
		state    types.String
		want     types.String
	}{
		{
			name:     "API returns value — use it",
			apiValue: "/projects/p/vpcs/new",
			state:    stateVal,
			want:     types.StringValue("/projects/p/vpcs/new"),
		},
		{
			name:     "API returns empty — fall back to state",
			apiValue: "",
			state:    stateVal,
			want:     stateVal,
		},
		{
			name:     "API returns empty, state is null — stay null",
			apiValue: "",
			state:    types.StringNull(),
			want:     types.StringNull(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveAPIStringRef(tc.apiValue, tc.state)
			if !got.Equal(tc.want) {
				t.Errorf("resolveAPIStringRef(%q, %v) = %v; want %v", tc.apiValue, tc.state, got, tc.want)
			}
		})
	}
}

func TestResolveKeyPairUriRef(t *testing.T) {
	stateWithKeypair := types.StringValue("/projects/p/keypairs/k")

	tests := []struct {
		name   string
		apiURI string
		state  types.String
		want   types.String
	}{
		{
			name:   "API returns URI — use it",
			apiURI: "/projects/p/keypairs/new",
			state:  stateWithKeypair,
			want:   types.StringValue("/projects/p/keypairs/new"),
		},
		{
			name:   "API returns empty, state has keypair — null (detached outside Terraform)",
			apiURI: "",
			state:  stateWithKeypair,
			want:   types.StringNull(),
		},
		{
			name:   "API returns empty, state has no keypair — preserve state",
			apiURI: "",
			state:  types.StringNull(),
			want:   types.StringNull(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveKeyPairUriRef(tc.apiURI, tc.state)
			if !got.Equal(tc.want) {
				t.Errorf("resolveKeyPairUriRef(%q, %v) = %v; want %v", tc.apiURI, tc.state, got, tc.want)
			}
		})
	}
}
