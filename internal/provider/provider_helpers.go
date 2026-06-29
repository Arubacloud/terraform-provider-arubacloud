package provider

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// effectiveTimeout returns the resource-level timeout if set and valid,
// otherwise falls back to the provider-level timeout.
func effectiveTimeout(resourceTimeout types.String, providerTimeout time.Duration) time.Duration {
	if resourceTimeout.IsNull() || resourceTimeout.IsUnknown() || resourceTimeout.ValueString() == "" {
		return providerTimeout
	}
	d, err := time.ParseDuration(resourceTimeout.ValueString())
	if err != nil || d <= 0 {
		return providerTimeout
	}
	return d
}
