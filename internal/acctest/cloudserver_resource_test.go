package acctest

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/terraform-provider-arubacloud/internal/provider"
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
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
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
				ImportStateIdFunc:       ImportIDFromAttrs("arubacloud_cloudserver.test", "project_id", "id"),
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
	client, err := AccClient()
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
		if provErr := provider.CheckResponseErr("get", "Cloudserver", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("CloudServer %s still exists", rs.Primary.ID)
	}
	return nil
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
