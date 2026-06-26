package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// compositeImportIDs maps resource names to their required composite import ID.
// All resources except "project" require at least project_id/resource_id.
var compositeImportIDs = map[string]string{
	// 2-part: project_id/resource_id
	"cloudserver":       "test-proj/test-srv",
	"backup":            "test-proj/test-bkp",
	"vpc":               "test-proj/test-vpc",
	"blockstorage":      "test-proj/test-vol",
	"snapshot":          "test-proj/test-snap",
	"dbaas":             "test-proj/test-dbaas",
	"databasebackup":    "test-proj/test-dbbackup",
	"kaas":              "test-proj/test-kaas",
	"containerregistry": "test-proj/test-reg",
	"keypair":           "test-proj/test-kp",
	"schedulejob":       "test-proj/test-job",
	"kms":               "test-proj/test-kms",
	"elasticip":         "test-proj/test-eip",
	"vpntunnel":         "test-proj/test-tun",
	// 3-part
	"restore":       "test-proj/test-bkp/test-rst",
	"subnet":        "test-proj/test-vpc/test-sub",
	"securitygroup": "test-proj/test-vpc/test-sg",
	"vpcpeering":    "test-proj/test-vpc/test-peer",
	"database":      "test-proj/test-dbaas/test-db",
	"dbaasuser":     "test-proj/test-dbaas/test-usr",
	"vpnroute":      "test-proj/test-tun/test-rte",
	// 4-part
	"securityrule":    "test-proj/test-vpc/test-sg/test-rule",
	"vpcpeeringroute": "test-proj/test-vpc/test-peer/test-rte",
	"databasegrant":   "test-proj/test-dbaas/test-db/test-usr",
	// "project" has no project_id field — keeps the default "test-import-id"
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
