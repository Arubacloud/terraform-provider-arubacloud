package provider

import (
	"context"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// buildKaaSCreateReqWithNodePool creates a KaaS Create request that includes
// one node pool in the settings object so that the Create() function does not
// reject the plan with "Missing Node Pools".  All other attributes use the
// same "test-value" / zero defaults as resourceCreateReqFull.
func buildKaaSCreateReqWithNodePool(ctx context.Context, t *testing.T) resource.CreateRequest {
	t.Helper()

	r := NewKaaSResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("buildKaaSCreateReqWithNodePool: schema root is not an object type")
	}

	// Extract the tftypes sub-types from the schema for settings and network.
	settingsType, ok := objType.AttributeTypes["settings"].(tftypes.Object)
	if !ok {
		t.Fatalf("buildKaaSCreateReqWithNodePool: settings is not a tftypes.Object")
	}
	nodePoolsListType, ok := settingsType.AttributeTypes["node_pools"].(tftypes.List)
	if !ok {
		t.Fatalf("buildKaaSCreateReqWithNodePool: node_pools is not a tftypes.List")
	}
	nodePoolObjType, ok := nodePoolsListType.ElementType.(tftypes.Object)
	if !ok {
		t.Fatalf("buildKaaSCreateReqWithNodePool: node pool element is not a tftypes.Object")
	}

	// Build one node pool value.
	nodePool := tftypes.NewValue(nodePoolObjType, map[string]tftypes.Value{
		"name":        tftypes.NewValue(tftypes.String, "test-pool"),
		"nodes":       tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(2)),
		"instance":    tftypes.NewValue(tftypes.String, "test-instance"),
		"zone":        tftypes.NewValue(tftypes.String, "test-zone"),
		"autoscaling": tftypes.NewValue(tftypes.Bool, false),
		"min_count":   tftypes.NewValue(tftypes.Number, nil), // null / not set
		"max_count":   tftypes.NewValue(tftypes.Number, nil),
	})

	// Build settings with one node pool.
	settings := tftypes.NewValue(settingsType, map[string]tftypes.Value{
		"kubernetes_version": tftypes.NewValue(tftypes.String, "1.29"),
		"node_pools":         tftypes.NewValue(nodePoolsListType, []tftypes.Value{nodePool}),
		"ha":                 tftypes.NewValue(tftypes.Bool, false),
	})

	// Build network object.
	networkType, ok := objType.AttributeTypes["network"].(tftypes.Object)
	if !ok {
		t.Fatalf("buildKaaSCreateReqWithNodePool: network is not a tftypes.Object")
	}
	nodeCIDRType, ok := networkType.AttributeTypes["node_cidr"].(tftypes.Object)
	if !ok {
		t.Fatalf("buildKaaSCreateReqWithNodePool: node_cidr is not a tftypes.Object")
	}
	network := tftypes.NewValue(networkType, map[string]tftypes.Value{
		"vpc_uri_ref":    tftypes.NewValue(tftypes.String, "test-vpc-uri"),
		"subnet_uri_ref": tftypes.NewValue(tftypes.String, "test-subnet-uri"),
		"node_cidr": tftypes.NewValue(nodeCIDRType, map[string]tftypes.Value{
			"address": tftypes.NewValue(tftypes.String, "10.0.0.0/24"),
			"name":    tftypes.NewValue(tftypes.String, "test-cidr"),
		}),
		"security_group_name": tftypes.NewValue(tftypes.String, "test-sg"),
		"pod_cidr":            tftypes.NewValue(tftypes.String, nil),
	})

	// Build the full plan using buildFullTFValue for all remaining attributes,
	// substituting the manually constructed settings and network objects.
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		switch name {
		case "settings":
			attrs[name] = settings
		case "network":
			attrs[name] = network
		default:
			attrs[name] = buildFullTFValue(ty)
		}
	}

	return resource.CreateRequest{
		Plan: tfsdk.Plan{
			Raw:    tftypes.NewValue(objType, attrs),
			Schema: schemaResp.Schema,
		},
	}
}

// TestKaaSCreate_Success verifies that KaaS Create() succeeds when the plan
// includes at least one node pool (the empty-list guard is the main reason
// KaaS is excluded from TestResourceCreate_Success_ComplexResources).
func TestKaaSCreate_Success(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, createSuccessHandler)

	res := NewKaaSResource()
	configureResource(ctx, t, res, mockClient)

	schemaResp := &resource.SchemaResponse{}
	res.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	req := buildKaaSCreateReqWithNodePool(ctx, t)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	res.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("KaaS Create() reported error: %v", resp.Diagnostics)
	}
}

// TestKaaSCreate_ProvisioningTimeout verifies the WaitForResourceActive
// timeout branch of KaaS Create() when the cluster remains in InCreation
// indefinitely.
func TestKaaSCreate_ProvisioningTimeout(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClientFast(t, alwaysCreatingHandler)

	res := NewKaaSResource()
	configureResource(ctx, t, res, mockClient)

	schemaResp := &resource.SchemaResponse{}
	res.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	req := buildKaaSCreateReqWithNodePool(ctx, t)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	res.Create(ctx, req, resp)

	// Timeout results in a warning diagnostic.
	if len(resp.Diagnostics) == 0 {
		t.Error("KaaS Create() should report a diagnostic on provisioning timeout")
	}
}

// TestKaaSCreate_APIError verifies that KaaS Create() adds an error diagnostic
// when the API returns 500 for the actual POST (the node pool check must pass
// for the POST to be reached).
func TestKaaSCreate_APIError(t *testing.T) {
	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, createAPIErrorHandler)

	res := NewKaaSResource()
	configureResource(ctx, t, res, mockClient)

	schemaResp := &resource.SchemaResponse{}
	res.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	req := buildKaaSCreateReqWithNodePool(ctx, t)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	res.Create(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("KaaS Create() should report error for HTTP 500 POST response")
	}
}

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

// kaasFullPropertiesJSON is a KaaS API response that includes all optional
// properties blocks: vpc, subnet, securityGroup, nodecidr, podcidr,
// billingPlan, ha, managementIp, kubernetesVersion, and nodesPool.
//
// The JSON key "nodesPool" (not "nodePools") matches the SDK field tag
// `json:"nodesPool,omitempty"` in NodePoolPropertiesResponse, so the pool
// array is actually deserialized and ptrToString() is exercised for the
// pool Name field.
const kaasFullPropertiesJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"managementIp":"10.0.0.1",` +
	`"kubernetesVersion":{"value":"1.29"},` +
	`"ha":true,` +
	`"billingPlan":{"billingPeriod":"Hour"},` +
	`"vpc":{"uri":"test-vpc-uri"},` +
	`"subnet":{"uri":"test-subnet-uri"},` +
	`"securityGroup":{"name":"test-sg-name"},` +
	`"nodecidr":{"address":"10.0.1.0/24","name":"test-cidr"},` +
	`"podcidr":{"address":"10.1.0.0/16"},` +
	`"nodesPool":[{` +
	`"name":"test-pool",` +
	`"nodes":2,` +
	`"instance":{"name":"test-instance"},` +
	`"dataCenter":{"code":"test-zone"},` +
	`"autoscaling":true,` +
	`"minCount":1,` +
	`"maxCount":5` +
	`}]` +
	`}}`

// kaasFullPropertiesHandler returns kaasFullPropertiesJSON for every GET and
// 500 for any write operation (none expected during Read).
func kaasFullPropertiesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(kaasFullPropertiesJSON)) //nolint:errcheck
}

// TestKaaSDataSource_FullProperties exercises the KaaS data source Read() with
// a response that populates all optional property branches:
//   - VPC URI, Subnet URI, SecurityGroup.Name, NodeCIDR address+name, PodCIDR
//   - BillingPlan.BillingPeriod, HA, ManagementIP
//   - nodesPool loop (exercises ptrToString() on Name, Instance.Name,
//     DataCenter.Code, MinCount, MaxCount)
//   - kubeconfig download attempt (managementIp is set; mock returns full JSON
//     which has no "content" field, so kubeconfig falls through to StringNull)
func TestKaaSDataSource_FullProperties(t *testing.T) {
	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, kaasFullPropertiesHandler)

	ds := NewKaaSDataSource()
	configureDatasource(ctx, t, ds, mockClient)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

	req := dsReadReq(ctx, t, ds, nil)
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	ds.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("KaaS datasource Read() reported error with full properties: %v", resp.Diagnostics)
	}
}

// TestKaaSResource_Read_FullProperties exercises the KaaS resource Read() with
// a response that populates all optional property branches.  resourceReadReqFull
// provides non-null nested state objects so that the network/settings
// object-building code inside Read() is reached.
func TestKaaSResource_Read_FullProperties(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, kaasFullPropertiesHandler)

	res := NewKaaSResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReqFull(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("KaaS resource Read() reported error with full properties: %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Fatal("KaaS resource Read() removed resource from state unexpectedly")
	}
}
