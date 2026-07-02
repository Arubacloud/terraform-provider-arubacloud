package provider

import (
	"context"
	"net/http"
	"testing"
)

// schedulejobWithStepsJSON is a ScheduleJob API response that includes a
// properties block with a non-empty steps array.  This exercises the step-
// mapping loop and the "if len(job.Properties.Steps) > 0" branch that is
// skipped when minimalActiveJSON (no properties) is used.
const schedulejobWithStepsJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"scheduleJobType":"OneShot",` +
	`"enabled":true,` +
	`"scheduleAt":"2099-01-01T00:00:00Z",` +
	`"steps":[{` +
	`"name":"test-step",` +
	`"resourceUri":"/projects/p/resources/r",` +
	`"actionUri":"/actions/a",` +
	`"httpVerb":"GET",` +
	`"body":"{}"` +
	`}]` +
	`}}`

// TestScheduleJobRead_WithSteps verifies that the Read() function correctly
// maps a response that includes a non-empty steps array, covering the
// step-mapping loop and ptrToString-style pointer branches inside it.
func TestScheduleJobRead_WithSteps(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(schedulejobWithStepsJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewScheduleJobResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ScheduleJob Read() reported error with steps: %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Error("ScheduleJob Read() removed resource from state with valid response")
	}
}

// TestScheduleJobRead_WithCronSchedule verifies the Read() function when the
// response includes cron, execute_until, and other optional scheduling fields.
func TestScheduleJobRead_WithCronSchedule(t *testing.T) {
	ctx := context.Background()

	cronJSON := `{` +
		`"metadata":{"id":"test-id","name":"test-name","location":{"value":"test-loc"},"tags":["env:test"]},` +
		`"status":{"state":"Active"},` +
		`"properties":{` +
		`"scheduleJobType":"Recurring",` +
		`"enabled":true,` +
		`"cron":"0 * * * *",` +
		`"executeUntil":"2099-12-31T23:59:59Z",` +
		`"steps":[]` +
		`}}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cronJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewScheduleJobResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ScheduleJob Read() reported error with cron: %v", resp.Diagnostics)
	}
}

// TestRestoreRead_WithProperties covers restore Read() with a response that
// includes a properties block to exercise property-mapping branches.
func TestRestoreRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	restoreJSON := `{` +
		`"metadata":{"id":"test-id","name":"test-name"},` +
		`"status":{"state":"Active"},` +
		`"properties":{` +
		`"billingPeriod":"Hour",` +
		`"origin":{"uri":"/backups/test-backup-id"}` +
		`}}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(restoreJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewRestoreResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Restore Read() reported error with properties: %v", resp.Diagnostics)
	}
}

// TestVPCPeeringRouteRead_WithProperties covers vpcpeeringroute Read() with
// a response that includes properties to exercise the mapping branches.
func TestVPCPeeringRouteRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	peerRouteJSON := `{` +
		`"metadata":{"id":"test-id","name":"test-name"},` +
		`"status":{"state":"Active"},` +
		`"properties":{` +
		`"route":{"destination":"10.1.0.0/24","gateway":"10.0.0.1"}` +
		`}}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(peerRouteJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewVpcPeeringRouteResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("VPCPeeringRoute Read() reported error with properties: %v", resp.Diagnostics)
	}
}

// TestVPNRouteRead_WithProperties covers vpnroute Read() with a properties
// block to exercise the property mapping.
func TestVPNRouteRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	vpnRouteJSON := `{` +
		`"metadata":{"id":"test-id","name":"test-name"},` +
		`"status":{"state":"Active"},` +
		`"properties":{` +
		`"cloudSubnet":"10.0.0.0/24",` +
		`"onPremSubnet":"192.168.0.0/24"` +
		`}}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(vpnRouteJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewVPNRouteResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReqFull(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("VPNRoute Read() reported error with properties: %v", resp.Diagnostics)
	}
}

// TestVPNTunnelRead_WithProperties covers vpntunnel Read() with a properties
// block to exercise the mapping of tunnel-specific fields.
func TestVPNTunnelRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	// vpntunnel has complex properties
	tunnelJSON := `{` +
		`"metadata":{"id":"test-id","name":"test-name"},` +
		`"status":{"state":"Active"},` +
		`"properties":{` +
		`"peerGatewayIP":"203.0.113.1",` +
		`"psk":"test-psk"` +
		`}}`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(tunnelJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewVPNTunnelResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReqFull(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("VPNTunnel Read() reported error with properties: %v", resp.Diagnostics)
	}
}
