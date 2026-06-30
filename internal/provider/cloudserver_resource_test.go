package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCloudserverResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	osImageID := os.Getenv("ARUBACLOUD_OS_IMAGE_ID")
	if projectID == "" || osImageID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_OS_IMAGE_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckCloudserverDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCloudserverResourceConfig(projectID, osImageID, "test-cloudserver"),
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
				ResourceName:            "arubacloud_cloudserver.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"settings.user_data", "network.securitygroup_uri_refs", "network.subnet_uri_refs"},
				ImportStateIdFunc:       importIDFromAttrs("arubacloud_cloudserver.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccCloudserverResourceConfig(projectID, osImageID, "test-cloudserver-updated"),
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
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/compute/cloudServers/" + rs.Primary.ID)
		_, err = client.Client.FromCompute().CloudServers().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Cloudserver", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
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

func testAccCloudserverResourceConfig(projectID, osImageID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "cs_prereq" {
  name       = "test-acc-cs-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "cs_prereq" {
  name       = "test-acc-cs-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.cs_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "cs_prereq" {
  name       = "test-acc-cs-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.cs_prereq.id
}

resource "arubacloud_blockstorage" "cs_boot" {
  name           = "test-acc-cs-boot"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 30
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
  bootable       = true
  image          = %[2]q
}

resource "arubacloud_cloudserver" "test" {
  name       = %[3]q
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  zone       = "ITBG-1"

  network = {
    vpc_uri_ref            = arubacloud_vpc.cs_prereq.uri
    subnet_uri_refs        = [arubacloud_subnet.cs_prereq.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.cs_prereq.uri]
  }

  settings = {
    flavor_name = "CSO4A8"
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.cs_boot.uri
  }
}
`, projectID, osImageID, name)
}
