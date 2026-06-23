package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestResourceImportState_Execute actually invokes ImportState() for all 25
// resources.  This covers ImportStatePassthroughID (the standard one-liner
// that all resources in this provider use) which the existing interface-only
// TestResourceImportState does not exercise.
func TestResourceImportState_Execute(t *testing.T) {
	ctx := context.Background()

	for _, tc := range allResources25 {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res := tc.newR()
			importable, ok := res.(resource.ResourceWithImportState)
			if !ok {
				t.Skipf("%s does not implement ResourceWithImportState", tc.name)
				return
			}

			schemaResp := &resource.SchemaResponse{}
			res.Schema(ctx, resource.SchemaRequest{}, schemaResp)

			// Build a typed-null raw state so ImportStatePassthroughID can
			// write the "id" attribute into it via SetAttribute.
			objType, ok2 := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
			if !ok2 {
				t.Fatalf("%s: schema root is not an object type", tc.name)
			}

			req := resource.ImportStateRequest{ID: "test-import-id"}
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
