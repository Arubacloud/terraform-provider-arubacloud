package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// useStateIfConfigNull is a string plan modifier that preserves the prior state
// value when the configuration omits the attribute (config value is null).
//
// Without this modifier, Optional+Computed attributes cause a perpetual plan diff
// when the API always returns a value (e.g. billing_period="Hour") but the user's
// config does not set the attribute: the Terraform Plugin Framework uses the
// config null as the planned value, overwriting the state on every refresh.
type useStateIfConfigNull struct{}

var _ planmodifier.String = useStateIfConfigNull{}

func (m useStateIfConfigNull) Description(_ context.Context) string {
	return "Use the prior state value when the configuration does not set the attribute."
}

func (m useStateIfConfigNull) MarkdownDescription(_ context.Context) string {
	return "Use the prior state value when the configuration does not set the attribute."
}

func (m useStateIfConfigNull) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return // config explicitly sets a value; use it
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return // nothing in state to preserve
	}
	resp.PlanValue = req.StateValue
}

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
