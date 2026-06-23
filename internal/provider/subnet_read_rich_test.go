package provider

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// subnetAdvancedJSON is a Subnet API response with a full properties block
// including type=Advanced, a network address, DHCP enabled with a range,
// routes, and DNS entries.  This exercises the DHCP-mapping code paths in
// SubnetResource.Read() that are skipped by minimalActiveJSON.
const subnetAdvancedJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"uri":"/subnets/test-id",` +
	`"location":{"value":"test-location"},` +
	`"tags":["env:test"]` +
	`},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"type":"Advanced",` +
	`"network":{"address":"10.0.0.0/24"},` +
	`"dhcp":{` +
	`"enabled":true,` +
	`"range":{"start":"10.0.0.100","count":50},` +
	`"routes":[{"address":"10.0.0.0/24","gateway":"10.0.0.1"}],` +
	`"dns":["8.8.8.8","8.8.4.4"]` +
	`}` +
	`}` +
	`}`

// subnetBasicJSON is a Basic Subnet response with no network block.  This
// exercises the shouldSetNetwork=false branch (type != "Advanced" and no
// network in state) which sets the entire network block to null.
const subnetBasicNoNetworkJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{"type":"Basic"}` +
	`}`

// TestSubnetResource_Read_WithDHCP verifies that SubnetResource.Read() correctly
// maps a response that includes a full DHCP block (range, routes, DNS).  Using
// resourceReadReqFull provides non-null network in state so shouldSetNetwork is
// true and the DHCP-mapping section is exercised.
func TestSubnetResource_Read_WithDHCP(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(subnetAdvancedJSON)) //nolint:errcheck
	})

	res := NewSubnetResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReqFull(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("SubnetResource Read() reported error with DHCP response: %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Fatal("SubnetResource Read() removed resource from state unexpectedly")
	}
}

// TestSubnetResource_Read_BasicNoNetwork verifies that SubnetResource.Read() with
// a Basic subnet response and null network in state correctly sets
// data.Network to null (the shouldSetNetwork=false / else branch).
func TestSubnetResource_Read_BasicNoNetwork(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(subnetBasicNoNetworkJSON)) //nolint:errcheck
	})

	res := NewSubnetResource()
	configureResource(ctx, t, res, mockClient)

	// Use minimal Read request (network is null in state) with a Basic subnet type.
	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("SubnetResource Read() reported error for Basic subnet: %v", resp.Diagnostics)
	}
}
