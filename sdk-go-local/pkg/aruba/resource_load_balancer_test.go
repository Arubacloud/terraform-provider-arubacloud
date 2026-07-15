package aruba

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time Ref satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*LoadBalancer)(nil)

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func loadBalancerTestResponse(id, name, uri, projectID string) *types.LoadBalancerResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	state := types.State("Active")
	addr := "10.0.0.1"
	vpcURI := "/projects/" + projectID + "/providers/Aruba.Network/vpcs/vpc-1"
	return &types.LoadBalancerResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"lb-tag"},
			LocationResponse: loc,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: projectID,
			},
		},
		Properties: types.LoadBalancerPropertiesResponse{
			Address: &addr,
			VPC:     &types.ReferenceResourceCommon{URI: vpcURI},
			LinkedResources: []types.LinkedResourceCommon{
				{URI: "/projects/p/providers/Aruba.Compute/cloudservers/cs1", StrictCorrelation: true},
			},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
			DisableStatusInfoResponse: &types.DisableStatusInfoResponse{
				IsDisabled: false,
			},
		},
	}
}

func TestLoadBalancer_FromResponseHydration(t *testing.T) {
	lb := &LoadBalancer{}
	resp := loadBalancerTestResponse("lb-1", "my-lb", "/projects/p1/providers/Aruba.Network/loadBalancers/lb-1", "p1")
	lb.fromResponse(resp)

	if lb.ID() != "lb-1" {
		t.Errorf("ID() = %q", lb.ID())
	}
	if lb.URI() != "/projects/p1/providers/Aruba.Network/loadBalancers/lb-1" {
		t.Errorf("URI() = %q", lb.URI())
	}
	if lb.LoadBalancerID() != "lb-1" {
		t.Errorf("LoadBalancerID() = %q", lb.LoadBalancerID())
	}
	if lb.Name() != "my-lb" {
		t.Errorf("Name() = %q", lb.Name())
	}
	if tags := lb.Tags(); len(tags) != 1 || tags[0] != "lb-tag" {
		t.Errorf("Tags() = %v", tags)
	}
	if lb.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", lb.Region())
	}
	if lb.State() != "Active" {
		t.Errorf("State() = %q", lb.State())
	}
	if lb.IsDisabled() {
		t.Error("IsDisabled() should be false")
	}
	if linked := lb.LinkedResources(); len(linked) != 1 {
		t.Errorf("LinkedResources() len = %d", len(linked))
	}
	if lb.Address() != "10.0.0.1" {
		t.Errorf("Address() = %q", lb.Address())
	}
	if lb.VPC() != "/projects/p1/providers/Aruba.Network/vpcs/vpc-1" {
		t.Errorf("VPC() = %q", lb.VPC())
	}
	if lb.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", lb.ProjectID())
	}
	if lb.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestLoadBalancer_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	lb := &LoadBalancer{}
	lb.fromResponse(nil)
	if lb.ID() != "" || lb.URI() != "" || lb.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}
	if lb.Raw() != nil {
		t.Error("Raw() should be nil before hydration")
	}

	// empty response — accessors must not panic; zero values expected
	lb2 := &LoadBalancer{}
	lb2.fromResponse(&types.LoadBalancerResponse{})
	if lb2.ID() != "" || lb2.URI() != "" || lb2.State() != "" || lb2.Address() != "" || lb2.VPC() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

func TestLoadBalancer_FromResponseProjectIDFromURI(t *testing.T) {
	uri := "/projects/p2/providers/Aruba.Network/loadBalancers/lb-2"
	id := "lb-2"
	name := "uri-lb"
	resp := &types.LoadBalancerResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			URI:  &uri,
			Name: &name,
			// ProjectMetadataResponse intentionally nil
		},
	}
	lb := &LoadBalancer{}
	lb.fromResponse(resp)

	if lb.ProjectID() != "p2" {
		t.Errorf("ProjectID() via URI fallback = %q, want %q", lb.ProjectID(), "p2")
	}
}

func TestLoadBalancer_RawNilUntilHydrated(t *testing.T) {
	lb := &LoadBalancer{}
	if lb.Raw() != nil {
		t.Error("Raw() must be nil before fromResponse is called")
	}
	resp := loadBalancerTestResponse("lb-1", "n", "/projects/p/providers/Aruba.Network/loadBalancers/lb-1", "p")
	lb.fromResponse(resp)
	if lb.Raw() != resp {
		t.Error("Raw() must return the exact resp pointer after fromResponse")
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestLoadBalancer_RefSatisfaction(t *testing.T) {
	lb := &LoadBalancer{}
	lb.fromResponse(loadBalancerTestResponse("lb-99", "n", "/projects/p99/providers/Aruba.Network/loadBalancers/lb-99", "p99"))

	// withLoadBalancerID typed path
	lid, ok := extractID(lb, func(r Ref) (string, bool) {
		if w, ok := r.(withLoadBalancerID); ok {
			return w.LoadBalancerID(), true
		}
		return "", false
	}, "loadBalancers")
	if !ok || lid != "lb-99" {
		t.Errorf("extractID via withLoadBalancerID = (%q, %v)", lid, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(lb, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", pid, ok)
	}
}

// --------------------------------------------------------------------------
// loadBalancerIDsFromRef helper
// --------------------------------------------------------------------------

func TestLoadBalancerIDsFromRef_TypedRef(t *testing.T) {
	lb := &LoadBalancer{}
	lb.fromResponse(loadBalancerTestResponse("lb-1", "n", "/projects/p/providers/Aruba.Network/loadBalancers/lb-1", "p"))
	pid, lid, err := loadBalancerIDsFromRef(lb)
	if err != nil || pid != "p" || lid != "lb-1" {
		t.Errorf("loadBalancerIDsFromRef typed = (%q, %q, %v)", pid, lid, err)
	}
}

func TestLoadBalancerIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Network/loadBalancers/lb-1")
	pid, lid, err := loadBalancerIDsFromRef(ref)
	if err != nil || pid != "p" || lid != "lb-1" {
		t.Errorf("loadBalancerIDsFromRef URI = (%q, %q, %v)", pid, lid, err)
	}
}

func TestLoadBalancerIDsFromRef_BadURI(t *testing.T) {
	_, _, err := loadBalancerIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for URI without /loadBalancers/<id>")
	}
}

func TestLoadBalancerIDsFromRef_NoProject(t *testing.T) {
	_, _, err := loadBalancerIDsFromRef(URI("/loadBalancers/lb-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// loadBalancersClientAdapter — integration tests
// --------------------------------------------------------------------------

func buildLoadBalancerTestAdapter(t *testing.T, handler http.HandlerFunc) *loadBalancersClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newLoadBalancersClientAdapter(testutil.NewClient(t, server.URL))
}

const loadBalancerSuccessBody = `{` +
	`"metadata":{"id":"lb-1","name":"my-lb","uri":"/projects/p/providers/Aruba.Network/loadBalancers/lb-1","project":{"id":"p"}},` +
	`"properties":{"address":"10.0.0.1"},` +
	`"status":{"state":"Active"}}`

func TestLoadBalancersClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, loadBalancerSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Network/loadBalancers/lb-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "lb-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if result.Address() != "10.0.0.1" {
		t.Errorf("Address() = %q", result.Address())
	}
	if result.StatusCode() != http.StatusOK {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	wantPath := "/projects/p/providers/Aruba.Network/loadBalancers/lb-1"
	if capturedPath != wantPath {
		t.Errorf("path = %q, want %q", capturedPath, wantPath)
	}
}

func TestLoadBalancersClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, loadBalancerSuccessBody)
	})

	existing := &LoadBalancer{}
	existing.fromResponse(loadBalancerTestResponse("lb-1", "n", "/projects/p/providers/Aruba.Network/loadBalancers/lb-1", "p"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "lb-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestLoadBalancersClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "load balancer not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Network/loadBalancers/lb-missing")
	result, err := adapter.Get(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

func TestLoadBalancersClientAdapter_Get_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	result, err := adapter.Get(context.Background(), URI("/garbage"))
	if err == nil {
		t.Fatal("expected error for unresolvable Ref")
	}
	if result != nil {
		t.Error("result should be nil on bad Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad Ref")
	}
}

func TestLoadBalancersClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"lb-1","name":"lb-a","uri":"/projects/p/providers/Aruba.Network/loadBalancers/lb-1","project":{"id":"p"}},"properties":{"address":"10.0.0.1"},"status":{"state":"Active"}},`+
			`{"metadata":{"id":"lb-2","name":"lb-b","uri":"/projects/p/providers/Aruba.Network/loadBalancers/lb-2","project":{"id":"p"}},"properties":{"address":"10.0.0.2"},"status":{"state":"Inactive"}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if list.Total() != 2 {
		t.Errorf("Total() = %d", list.Total())
	}
	items := list.Items()
	if len(items) != 2 {
		t.Fatalf("Items() len = %d", len(items))
	}
	if items[0].ID() != "lb-1" || items[0].Name() != "lb-a" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].Address() != "10.0.0.1" {
		t.Errorf("items[0].Address() = %q", items[0].Address())
	}
	if items[1].ID() != "lb-2" || items[1].State() != "Inactive" {
		t.Errorf("items[1] ID=%q State=%q", items[1].ID(), items[1].State())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

func TestLoadBalancersClientAdapter_List_Empty(t *testing.T) {
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":0,"self":"","prev":"","next":"","first":"","last":"","values":[]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if list.Total() != 0 {
		t.Errorf("Total() = %d", list.Total())
	}
	if len(list.Items()) != 0 {
		t.Errorf("Items() len = %d", len(list.Items()))
	}
}

func TestLoadBalancersClientAdapter_List_BadProjectRef(t *testing.T) {
	callCount := 0
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	result, err := adapter.List(context.Background(), URI("/garbage"))
	if err == nil {
		t.Fatal("expected error for unresolvable project Ref")
	}
	if result != nil {
		t.Error("result should be nil on bad project Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad project Ref")
	}
}

func TestLoadBalancersClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Internal Server Error", "unexpected error", 500))
	})

	result, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error on 500")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
	if result != nil {
		t.Error("result should be nil on non-2xx List")
	}
}

func TestLoadBalancersClientAdapter_Get_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newLoadBalancersClientAdapter(testutil.NewClient(t, server.URL))
	result, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/loadBalancers/lb-1"))
	if err == nil {
		t.Fatal("expected transport error")
	}
	_ = result
}

func TestLoadBalancersClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newLoadBalancersClientAdapter(testutil.NewClient(t, server.URL))
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestLoadBalancersClientAdapter_List_ProjectIDBackfill(t *testing.T) {
	// Response items without projectID in metadata or URI: triggers lb.projectID = projectID backfill
	adapter := buildLoadBalancerTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// No "project" key and no URI path that can yield a projectID
		fmt.Fprint(w, `{"total":1,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"lb-x","name":"lb-x"},"properties":{},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/proj-x"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	items := list.Items()
	if len(items) != 1 {
		t.Fatalf("Items() len = %d", len(items))
	}
	if items[0].ProjectID() != "proj-x" {
		t.Errorf("ProjectID() after backfill = %q, want %q", items[0].ProjectID(), "proj-x")
	}
}

func TestLoadBalancer_FromResponse_SetsStatus(t *testing.T) {
	l := &LoadBalancer{}
	state := types.State("Active")
	l.fromResponse(&types.LoadBalancerResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if l.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", l.State())
	}
}

func TestLoadBalancersClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, loadBalancerSuccessBody)
	})
	adapter := newLoadBalancersClientAdapter(testutil.NewClient(t, server.URL))
	lb, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/loadBalancers/lb-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&lb.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned LoadBalancer")
	}
}

func TestLoadBalancerRef(t *testing.T) {
	ref := LoadBalancerRef("p-1", "lb-1")
	want := "/projects/p-1/providers/Aruba.Network/loadBalancers/lb-1"
	if ref.URI() != want {
		t.Errorf("LoadBalancerRef URI = %q, want %q", ref.URI(), want)
	}
	ids := parseURIIDs(ref.URI())
	if ids["projects"] != "p-1" || ids["loadBalancers"] != "lb-1" {
		t.Errorf("parseURIIDs = %v", ids)
	}
}
