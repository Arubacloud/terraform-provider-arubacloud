package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TagsToList converts a []string of tags to a types.List suitable for storing
// in Terraform state. An empty or nil slice produces an empty list (never null).
func TagsToList(tags []string) types.List {
	values := make([]attr.Value, len(tags))
	for i, tag := range tags {
		values[i] = types.StringValue(tag)
	}
	return types.ListValueMust(types.StringType, values)
}

// TagsToListPreserveNull is like TagsToList but avoids a null→[] inconsistency
// when the user omitted the optional tags attribute. If the API returns no tags
// AND the prior state/plan value was null (user never set the attribute), null
// is returned so Terraform does not flag a phantom diff. If the prior value was
// an explicit list (empty or populated), the API value is used as-is.
func TagsToListPreserveNull(tags []string, prior types.List) types.List {
	if len(tags) == 0 && prior.IsNull() {
		return types.ListNull(types.StringType)
	}
	return TagsToList(tags)
}

// ListToTags converts a types.List of strings from Terraform state into a
// []string suitable for API requests. Null and unknown lists return nil without
// appending any diagnostics.
func ListToTags(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var tags []string
	diags.Append(list.ElementsAs(ctx, &tags, false)...)
	return tags
}

// billingPeriodFromAPI normalizes legacy API billing period values to their
// canonical Terraform form. The API has historically returned lowercase variants
// ("hourly", "monthly", "yearly") that differ from the accepted input values
// ("Hour", "Month", "Year"). Values already in canonical form pass through unchanged.
func billingPeriodFromAPI(s string) string {
	switch s {
	case "hourly":
		return "Hour"
	case "monthly":
		return "Month"
	case "yearly":
		return "Year"
	default:
		return s
	}
}
