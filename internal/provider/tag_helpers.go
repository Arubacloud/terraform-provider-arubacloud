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
