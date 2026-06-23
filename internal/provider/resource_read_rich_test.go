package provider

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// richMetadataJSON extends minimalActiveJSON with the additional metadata
// fields that most resource Read() functions handle in uncovered branches:
//   - metadata.uri      → the "URI non-nil" true branch (sets data.Uri)
//   - metadata.location → the "LocationResponse non-nil" true branch (sets data.Location)
//   - metadata.tags     → the "len(Tags)>0" true branch (maps tags list)
//
// These branches are skipped by minimalActiveJSON in TestResourceRead_Success,
// leaving 5-10 uncovered lines per resource.
const richMetadataJSON = `{` +
	`"metadata":{` +
	`"id":"test-id",` +
	`"name":"test-name",` +
	`"uri":"/projects/p/resources/test-id",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"},` +
	`"tags":["env:test","team:platform"]` +
	`},` +
	`"status":{"state":"Active"}` +
	`}`

// richSnapshotJSON adds a volume pointer field (required by snapshot Read()) on
// top of the standard rich metadata.
const richSnapshotJSON = `{` +
	`"metadata":{` +
	`"id":"test-id",` +
	`"name":"test-name",` +
	`"uri":"/projects/p/snapshots/test-id",` +
	`"location":{"value":"test-location"},` +
	`"tags":["env:test"]` +
	`},` +
	`"status":{"state":"Active"},` +
	`"properties":{"volume":{"uri":"/projects/p/volumes/test-vol-id"}}` +
	`}`

// richMetadataHandler returns richMetadataJSON for every GET.
func richMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(richMetadataJSON)) //nolint:errcheck
		return
	}
	apiError(w, http.StatusInternalServerError)
}

// TestResourceRead_RichMetadata re-runs the same set of resources as
// TestResourceRead_Success but with richMetadataJSON instead of
// minimalActiveJSON.  This covers the metadata.URI, metadata.LocationResponse,
// and metadata.Tags branches that minimalActiveJSON leaves as uncovered "false"
// paths across all 21 resources in resourcesForReadSuccess.
func TestResourceRead_RichMetadata(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	for _, tc := range resourcesForReadSuccess {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := richMetadataHandler
			if tc.name == "snapshot" {
				handler = func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodGet {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(richSnapshotJSON)) //nolint:errcheck
						return
					}
					apiError(w, http.StatusInternalServerError)
				}
			}

			_, mockClient := newMockArubaClient(t, handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() reported error with rich metadata: %v",
					tc.name, resp.Diagnostics)
			}
			if resp.State.Raw.IsNull() {
				t.Errorf("%s: Read() removed resource from state with rich metadata", tc.name)
			}
		})
	}
}

// TestDataSourceRead_RichMetadata runs all 25 datasources with richMetadataJSON
// to cover the URI / location / tags metadata branches in datasource Read().
func TestDataSourceRead_RichMetadata(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name    string
		newDS   func() datasource.DataSource
		handler http.HandlerFunc
	}{
		{"vpc", NewVPCDataSource, richMetadataHandler},
		{"subnet", NewSubnetDataSource, richMetadataHandler},
		{"securitygroup", NewSecurityGroupDataSource, richMetadataHandler},
		{"securityrule", NewSecurityRuleDataSource, richMetadataHandler},
		{"elasticip", NewElasticIPDataSource, richMetadataHandler},
		{"keypair", NewKeypairDataSource, richMetadataHandler},
		{"blockstorage", NewBlockStorageDataSource, richMetadataHandler},
		{"snapshot", NewSnapshotDataSource, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(richSnapshotJSON)) //nolint:errcheck
		}},
		{"backup", NewBackupDataSource, richMetadataHandler},
		{"restore", NewRestoreDataSource, richMetadataHandler},
		{"kms", NewKMSDataSource, richMetadataHandler},
		{"project", NewProjectDataSource, richMetadataHandler},
		{"cloudserver", NewCloudServerDataSource, richMetadataHandler},
		{"vpcpeering", NewVPCPeeringDataSource, richMetadataHandler},
		{"vpcpeeringroute", NewVPCPeeringRouteDataSource, richMetadataHandler},
		{"vpntunnel", NewVPNTunnelDataSource, richMetadataHandler},
		{"vpnroute", NewVPNRouteDataSource, richMetadataHandler},
		{"dbaas", NewDBaaSDataSource, richMetadataHandler},
		{"dbaasuser", NewDBaaSUserDataSource, richMetadataHandler},
		{"database", NewDatabaseDataSource, richMetadataHandler},
		{"databasebackup", NewDatabaseBackupDataSource, richMetadataHandler},
		{"databasegrant", NewDatabaseGrantDataSource, richMetadataHandler},
		{"kaas", NewKaaSDataSource, richMetadataHandler},
		{"containerregistry", NewContainerRegistryDataSource, containerRegistryReadSuccessHandler},
		{"schedulejob", NewScheduleJobDataSource, richMetadataHandler},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, tc.handler)

			ds := tc.newDS()
			configureDatasource(ctx, t, ds, mockClient)

			schemaResp := &datasource.SchemaResponse{}
			ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)

			req := dsReadReq(ctx, t, ds, nil)
			resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
			ds.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s datasource: Read() error with rich metadata: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}

// TestResourceRead_RichMetadata_ComplexResources runs the complex resources
// (cloudserver, containerregistry, dbaas, securityrule) with rich metadata
// and appropriate state/response fixtures so that URI/location/tags branches
// are covered in their Read() functions too.
func TestResourceRead_RichMetadata_ComplexResources(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	// cloudserver: needs non-null nested objects in state + rich metadata response
	t.Run("cloudserver", func(t *testing.T) {
		// cloudserverFullJSON already includes uri and location.
		_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Add tags to cloudserverFullJSON
			w.Write([]byte(`{"metadata":{"id":"test-id","name":"test-name",` +
				`"uri":"/cloudservers/test-id",` +
				`"location":{"value":"test-location"},` +
				`"tags":["env:test"]},` +
				`"status":{"state":"Active"},` +
				`"properties":{"vpc":{"uri":"test-vpc-uri"},` +
				`"bootVolume":{"uri":"test-boot-uri"},` +
				`"flavor":{"name":"test-flavor"},` +
				`"keyPair":{"uri":""},"zone":"test-zone"}}`)) //nolint:errcheck
		})
		res := NewCloudServerResource()
		configureResource(ctx, t, res, mockClient)
		req, resp := resourceReadReqFull(ctx, t, res)
		res.Read(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("cloudserver: Read() error with tags: %v", resp.Diagnostics)
		}
	})

	// securityrule: needs target in response + standard state
	t.Run("securityrule", func(t *testing.T) {
		richSecRuleJSON := `{"metadata":{"id":"test-id","name":"test-name",` +
			`"uri":"/securityrules/test-id",` +
			`"location":{"value":"test-location"},` +
			`"tags":["env:test"]},` +
			`"status":{"state":"Active"},` +
			`"properties":{"direction":"Ingress","protocol":"TCP","port":"80",` +
			`"target":{"kind":"IP","value":"10.0.0.0/8"}}}`
		_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(richSecRuleJSON)) //nolint:errcheck
		})
		res := NewSecurityRuleResource()
		configureResource(ctx, t, res, mockClient)
		req, resp := resourceReadReq(ctx, t, res)
		res.Read(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("securityrule: Read() error with tags+location: %v", resp.Diagnostics)
		}
	})

	// containerregistry: needs Properties + rich metadata
	t.Run("containerregistry", func(t *testing.T) {
		richCRJSON := `{"metadata":{"id":"test-id","name":"test-name",` +
			`"uri":"/containerregistries/test-id",` +
			`"location":{"value":"test-location"},` +
			`"tags":["env:test"]},` +
			`"status":{"state":"Active"},` +
			`"properties":{"publicIp":{"uri":"test-pip-uri"},"vpc":{"uri":"test-vpc-uri"},` +
			`"subnet":{"uri":"test-subnet-uri"},"securityGroup":{"uri":"test-sg-uri"},` +
			`"blockStorage":{"uri":"test-bs-uri"},` +
			`"billingPlan":{"billingPeriod":"Month"},` +
			`"adminUser":{"username":"test-admin"}}}`
		_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(richCRJSON)) //nolint:errcheck
		})
		res := NewContainerRegistryResource()
		configureResource(ctx, t, res, mockClient)
		req, resp := resourceReadReqFull(ctx, t, res)
		res.Read(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("containerregistry: Read() error with tags+location: %v", resp.Diagnostics)
		}
	})

	// dbaas: needs non-null nested objects in state
	t.Run("dbaas", func(t *testing.T) {
		_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(richMetadataJSON)) //nolint:errcheck
		})
		res := NewDBaaSResource()
		configureResource(ctx, t, res, mockClient)
		req, resp := resourceReadReqFull(ctx, t, res)
		res.Read(ctx, req, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("dbaas: Read() error with rich metadata: %v", resp.Diagnostics)
		}
	})
}
