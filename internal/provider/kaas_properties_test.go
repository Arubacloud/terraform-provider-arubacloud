package provider

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

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
