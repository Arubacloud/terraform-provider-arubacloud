package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// LogAndAppendAPIError emits a tflog.Error with the given context fields so the
// failure is visible under TF_LOG=DEBUG, then appends the error to diags as a
// user-facing diagnostic. `fields` should carry identifying IDs (project_id,
// vpc_id, etc.) for log correlation — err.Error() is added automatically under
// the "error" key.
func LogAndAppendAPIError(ctx context.Context, diags *diag.Diagnostics, summary string, err error, fields map[string]any) {
	logFields := make(map[string]any, len(fields)+1)
	for k, v := range fields {
		logFields[k] = v
	}
	logFields["error"] = err.Error()
	tflog.Error(ctx, summary, logFields)
	diags.AddError(summary, err.Error())
}
