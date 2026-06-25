package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// compositeImportIDs maps resource names to their required composite import ID
// format.  Resources not in this map use the simple "test-import-id" default.
var compositeImportIDs = map[string]string{
	// databasegrant requires project_id/dbaas_id/database/user_id
	"databasegrant": "test-project_id/test-dbaas_id/test-database/test-user_id",
}

// TestResourceImportState_Success invokes ImportState() for all 25 resources.
// This covers ImportStatePassthroughID (the standard one-liner every resource
// in this provider uses) which the existing interface-only TestResourceImportState
// in resource_schema_test.go does not execute.
func TestResourceImportState_Success(t *testing.T) {
	ctx := context.Background()

	for _, tc := range allResources25 {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.newR()
			importable, ok := res.(resource.ResourceWithImportState)
			if !ok {
				t.Skipf("%s does not implement ResourceWithImportState", tc.name)
				return
			}

			schemaResp := &resource.SchemaResponse{}
			res.Schema(ctx, resource.SchemaRequest{}, schemaResp)

			// Build a typed-null raw state so ImportStatePassthroughID can write
			// the "id" attribute into it via SetAttribute.
			objType, ok2 := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
			if !ok2 {
				t.Fatalf("%s: schema root is not an object type", tc.name)
			}

			importID := "test-import-id"
			if composite, exists := compositeImportIDs[tc.name]; exists {
				importID = composite
			}

			req := resource.ImportStateRequest{ID: importID}
			resp := &resource.ImportStateResponse{
				State: tfsdk.State{
					Raw:    tftypes.NewValue(objType, nil),
					Schema: schemaResp.Schema,
				},
			}
			importable.ImportState(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: ImportState() failed: %v", tc.name, resp.Diagnostics)
			}
		})
	}
}
