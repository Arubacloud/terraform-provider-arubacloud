package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCloudserverResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckCloudserverDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCloudserverResourceConfig("test-cloudserver"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-cloudserver"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("zone"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_cloudserver.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccCloudserverResourceConfig("test-cloudserver-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_cloudserver.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-cloudserver-updated"),
					),
				},
			},
		},
	})
}

func testCheckCloudserverDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_cloudserver" {
			continue
		}
		resp, err := client.Client.FromCompute().CloudServers().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Cloudserver", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("CloudServer %s still exists", rs.Primary.ID)
	}
	return nil
}

func TestResolveAPIStringRef(t *testing.T) {
	stateVal := types.StringValue("/projects/p/vpcs/v")

	tests := []struct {
		name     string
		apiValue string
		state    types.String
		want     types.String
	}{
		{
			name:     "API returns value — use it",
			apiValue: "/projects/p/vpcs/new",
			state:    stateVal,
			want:     types.StringValue("/projects/p/vpcs/new"),
		},
		{
			name:     "API returns empty — fall back to state",
			apiValue: "",
			state:    stateVal,
			want:     stateVal,
		},
		{
			name:     "API returns empty, state is null — stay null",
			apiValue: "",
			state:    types.StringNull(),
			want:     types.StringNull(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveAPIStringRef(tc.apiValue, tc.state)
			if !got.Equal(tc.want) {
				t.Errorf("resolveAPIStringRef(%q, %v) = %v; want %v", tc.apiValue, tc.state, got, tc.want)
			}
		})
	}
}

func TestResolveKeyPairUriRef(t *testing.T) {
	stateWithKeypair := types.StringValue("/projects/p/keypairs/k")

	tests := []struct {
		name   string
		apiURI string
		state  types.String
		want   types.String
	}{
		{
			name:   "API returns URI — use it",
			apiURI: "/projects/p/keypairs/new",
			state:  stateWithKeypair,
			want:   types.StringValue("/projects/p/keypairs/new"),
		},
		{
			name:   "API returns empty, state has keypair — null (detached outside Terraform)",
			apiURI: "",
			state:  stateWithKeypair,
			want:   types.StringNull(),
		},
		{
			name:   "API returns empty, state has no keypair — preserve state",
			apiURI: "",
			state:  types.StringNull(),
			want:   types.StringNull(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveKeyPairUriRef(tc.apiURI, tc.state)
			if !got.Equal(tc.want) {
				t.Errorf("resolveKeyPairUriRef(%q, %v) = %v; want %v", tc.apiURI, tc.state, got, tc.want)
			}
		})
	}
}

func testAccCloudserverResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_cloudserver" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
  zone       = "it-1"
  
  network = {
    vpc_uri_ref              = "test-vpc-uri"
    subnet_uri_refs          = ["test-subnet-uri"]
    securitygroup_uri_refs   = ["test-sg-uri"]
  }
  
  settings = {
    flavor_name = "test-flavor"
  }
  
  storage = {
    boot_volume_uri_ref = "test-volume-uri"
  }
}
`, name)
}
