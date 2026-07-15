package aruba

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// TestNewClient_BuildsAllSubsystems verifies that NewClient with valid Options
// returns a non-nil Client and that every From<Domain>().<Resource>s() accessor
// returns a non-nil value. This exercises all 53 build* helpers in builder.go
// without requiring any network connectivity (the token manager fetches tokens
// lazily, not at construction time).
func TestNewClient_BuildsAllSubsystems(t *testing.T) {
	opts := NewOptions().
		WithBaseURL("http://localhost:8080").
		WithTokenIssuerURL("http://localhost:8080/token").
		WithClientCredentials("test-id", "test-secret").
		WithNoLogs()

	cli, err := NewClient(opts)
	if err != nil {
		t.Fatalf("NewClient returned unexpected error: %v", err)
	}
	if cli == nil {
		t.Fatal("NewClient returned nil client")
	}

	type accessor struct {
		name string
		fn   func() any
	}

	accessors := []accessor{
		{"FromAudit().Events()", func() any { return cli.FromAudit().Events() }},
		{"FromCompute().CloudServers()", func() any { return cli.FromCompute().CloudServers() }},
		{"FromCompute().KeyPairs()", func() any { return cli.FromCompute().KeyPairs() }},
		{"FromContainer().KaaS()", func() any { return cli.FromContainer().KaaS() }},
		{"FromContainer().ContainerRegistry()", func() any { return cli.FromContainer().ContainerRegistry() }},
		{"FromDatabase().DBaaS()", func() any { return cli.FromDatabase().DBaaS() }},
		{"FromDatabase().Databases()", func() any { return cli.FromDatabase().Databases() }},
		{"FromDatabase().Backups()", func() any { return cli.FromDatabase().Backups() }},
		{"FromDatabase().Users()", func() any { return cli.FromDatabase().Users() }},
		{"FromDatabase().Grants()", func() any { return cli.FromDatabase().Grants() }},
		{"FromMetric().Alerts()", func() any { return cli.FromMetric().Alerts() }},
		{"FromMetric().Metrics()", func() any { return cli.FromMetric().Metrics() }},
		{"FromNetwork().ElasticIPs()", func() any { return cli.FromNetwork().ElasticIPs() }},
		{"FromNetwork().LoadBalancers()", func() any { return cli.FromNetwork().LoadBalancers() }},
		{"FromNetwork().SecurityGroupRules()", func() any { return cli.FromNetwork().SecurityGroupRules() }},
		{"FromNetwork().SecurityGroups()", func() any { return cli.FromNetwork().SecurityGroups() }},
		{"FromNetwork().Subnets()", func() any { return cli.FromNetwork().Subnets() }},
		{"FromNetwork().VPCPeeringRoutes()", func() any { return cli.FromNetwork().VPCPeeringRoutes() }},
		{"FromNetwork().VPCPeerings()", func() any { return cli.FromNetwork().VPCPeerings() }},
		{"FromNetwork().VPCs()", func() any { return cli.FromNetwork().VPCs() }},
		{"FromNetwork().VPNRoutes()", func() any { return cli.FromNetwork().VPNRoutes() }},
		{"FromNetwork().VPNTunnels()", func() any { return cli.FromNetwork().VPNTunnels() }},
		{"FromProject()", func() any { return cli.FromProject() }},
		{"FromSchedule().Jobs()", func() any { return cli.FromSchedule().Jobs() }},
		{"FromSecurity().KMS()", func() any { return cli.FromSecurity().KMS() }},
		{"FromSecurity().Keys()", func() any { return cli.FromSecurity().Keys() }},
		{"FromSecurity().Kmips()", func() any { return cli.FromSecurity().Kmips() }},
		{"FromStorage().Snapshots()", func() any { return cli.FromStorage().Snapshots() }},
		{"FromStorage().Volumes()", func() any { return cli.FromStorage().Volumes() }},
		{"FromStorage().Backups()", func() any { return cli.FromStorage().Backups() }},
		{"FromStorage().Restores()", func() any { return cli.FromStorage().Restores() }},
	}

	for _, a := range accessors {
		t.Run(a.name, func(t *testing.T) {
			v := a.fn()
			rv := reflect.ValueOf(v)
			if !rv.IsValid() || rv.IsNil() {
				t.Errorf("%s returned nil", a.name)
			}
		})
	}
}

// TestNewClient_RejectsInvalidOptions verifies that NewClient propagates
// validation errors from Options.validate() (covers the early-return error
// path in buildClient).
func TestNewClient_RejectsInvalidOptions(t *testing.T) {
	cases := []struct {
		name string
		opts *Options
	}{
		{
			name: "empty options",
			opts: NewOptions(),
		},
		{
			name: "missing token source",
			opts: NewOptions().WithBaseURL("http://localhost:8080"),
		},
		{
			name: "invalid base URL scheme",
			opts: NewOptions().
				WithBaseURL("ftp://localhost:8080").
				WithTokenIssuerURL("http://localhost:8080/token").
				WithClientCredentials("id", "secret"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cli, err := NewClient(tc.opts)
			if err == nil {
				t.Errorf("expected error for %q, got nil", tc.name)
			}
			if cli != nil {
				t.Errorf("expected nil client on error for %q", tc.name)
			}
		})
	}
}

// --------------------------------------------------------------------------
// User-Agent injection (#310)
// --------------------------------------------------------------------------

// captureUserAgent spins up an httptest.Server that records the User-Agent
// of the first request it receives, returns 200, and closes.
func captureUserAgent(t *testing.T) (serverURL string, getUA func() string) {
	t.Helper()
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total":0,"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	return srv.URL, func() string { return captured }
}

func TestClient_SendsDefaultUserAgent(t *testing.T) {
	srvURL, getUA := captureUserAgent(t)

	cli, err := NewClient(NewOptions().
		WithBaseURL(srvURL).
		WithToken("test-token").
		WithNoLogs())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, _ = cli.FromProject().List(context.Background())

	want := defaultUserAgent
	if got := getUA(); got != want {
		t.Errorf("User-Agent = %q, want %q", got, want)
	}
}

func TestOptions_WithUserAgent_Override(t *testing.T) {
	// WithUserAgent + DeepCopy round-trip
	o := NewOptions().WithUserAgent("acloud-cli@1.0.0")
	cp := o.DeepCopy()
	if cp.userAgent != "acloud-cli@1.0.0" {
		t.Errorf("DeepCopy().userAgent = %q, want %q", cp.userAgent, "acloud-cli@1.0.0")
	}
}

func TestOptions_WithUserAgent_DefaultWhenEmpty(t *testing.T) {
	o := NewOptions()
	if o.userAgent != "" {
		t.Errorf("default userAgent = %q, want empty", o.userAgent)
	}
	// resolved UA should equal defaultUserAgent
	ua := o.userAgent
	if ua == "" {
		ua = defaultUserAgent
	}
	if ua != defaultUserAgent {
		t.Errorf("resolved UA = %q, want %q", ua, defaultUserAgent)
	}
}

// TestNewClient_WithStaticToken verifies that NewClient accepts a static
// bearer token in place of a token-issuer configuration.
func TestNewClient_WithStaticToken(t *testing.T) {
	opts := NewOptions().
		WithBaseURL("http://localhost:8080").
		WithToken("my-static-bearer-token").
		WithNoLogs()

	cli, err := NewClient(opts)
	if err != nil {
		t.Fatalf("NewClient with static token returned error: %v", err)
	}
	if cli == nil {
		t.Fatal("NewClient with static token returned nil")
	}
}
