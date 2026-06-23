package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// kaasNodePoolJSON is a KaaS API response that includes a properties block
// with one node pool.  This exercises the ptrToString() helper, the node-pool
// loop, and the managementIP / kubeconfig-download branch inside the KaaS
// data source and resource Read() functions.
const kaasNodePoolJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"managementIP":"10.0.0.1",` +
	`"kubernetesVersion":{"value":"1.29"},` +
	`"nodePools":[{` +
	`"name":"test-pool",` +
	`"nodes":2,` +
	`"instance":{"name":"test-instance"},` +
	`"dataCenter":{"code":"test-zone"},` +
	`"autoscaling":false` +
	`}]` +
	`}}`

// TestKaaSDataSourceRead_WithNodePools verifies that the KaaS data source
// Read() correctly maps a response that includes node pools, covering the
// ptrToString() helper and the MinCount/MaxCount null branches.
func TestKaaSDataSourceRead_WithNodePools(t *testing.T) {
	ctx := context.Background()

	// For the kubeconfig download request the mock returns minimalActiveJSON
	// (no "content" field) so kubeconfigResp.Data.Content will be "" and the
	// function falls through to data.Kubeconfig = types.StringNull().
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(kaasNodePoolJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	ds := NewKaaSDataSource()
	configureDatasource(ctx, t, ds, mockClient)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

	req := dsReadReq(ctx, t, ds, nil)
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	ds.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("KaaS datasource Read() reported error with node pools: %v", resp.Diagnostics)
	}
}

// TestKaaSResourceRead_WithNodePools verifies that the KaaS resource Read()
// maps a response with node pools and managementIP correctly.
func TestKaaSResourceRead_WithNodePools(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(kaasNodePoolJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewKaaSResource()
	configureResource(ctx, t, res, mockClient)

	// The KaaS resource Read() accesses nested objects from the state.
	// Use resourceReadReqFull to provide non-null network and settings objects.
	req, resp := resourceReadReqFull(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("KaaS resource Read() reported error with node pools: %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Fatal("KaaS resource Read() removed resource from state unexpectedly")
	}
}
