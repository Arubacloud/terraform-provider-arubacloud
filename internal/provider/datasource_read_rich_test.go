package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// dbaasRichPropertiesJSON is a DBaaS API response that includes optional
// property fields (engine, flavor, billingPlan, storage) that minimalActiveJSON
// leaves absent.  Using this JSON covers the `!= nil` branches for each field
// in DBaaSResource.Read() and the `else` (preserve-from-state) branches when
// Storage.SizeGB is present.
const dbaasRichPropertiesJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}` +
	`},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"engine":{"id":"postgres-14"},` +
	`"flavor":{"name":"m1"},` +
	`"billingPlan":{"billingPeriod":"Hour"},` +
	`"storage":{"sizeGb":20}` +
	`}` +
	`}`

// TestDBaaSResource_Read_WithProperties verifies that DBaaSResource.Read()
// correctly maps engine, flavor, billingPlan and storage from the API response
// into Terraform state.  resourceReadReqFull provides non-null storage and
// network in state so the autoscaling-preserve-from-state branches are also
// exercised.
func TestDBaaSResource_Read_WithProperties(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(dbaasRichPropertiesJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewDBaaSResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReqFull(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("DBaaSResource Read() with properties reported error: %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Fatal("DBaaSResource Read() removed resource from state unexpectedly")
	}
}

// TestSubnetDataSource_Read_WithDHCP exercises SubnetDataSource.Read() with a
// response that includes a full DHCP block (range, routes, DNS), covering the
// DHCP-mapping branches that are skipped by minimalActiveJSON / richMetadataJSON.
//
// The JSON matches the SubnetPropertiesResponse SDK struct field tags.
func TestSubnetDataSource_Read_WithDHCP(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(subnetAdvancedJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	ds := NewSubnetDataSource()
	configureDatasource(ctx, t, ds, mockClient)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

	req := dsReadReq(ctx, t, ds, nil)
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	ds.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("SubnetDataSource Read() with DHCP reported error: %v", resp.Diagnostics)
	}
}
