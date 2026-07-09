package acctest

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCloudserverDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	osImageID := os.Getenv("ARUBACLOUD_OS_IMAGE_ID")
	if projectID == "" || osImageID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_OS_IMAGE_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudserverDataSourceConfig(projectID, osImageID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("flavor_name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_cloudserver.test",
						tfjsonpath.New("tags"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccCloudserverDataSourceConfig(projectID, osImageID string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "test" {
  name       = "test-ds-cs-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "test" {
  name       = "test-ds-cs-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "test" {
  name       = "test-ds-cs-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
}

resource "arubacloud_blockstorage" "boot" {
  name           = "test-ds-cs-boot"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 30
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
  bootable       = true
  image          = %[2]q
  timeout        = "2h"
}

resource "arubacloud_cloudserver" "test" {
  name       = "test-ds-cloudserver"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  zone       = "ITBG-1"
  tags       = ["acceptance-test"]

  network = {
    vpc_uri_ref            = arubacloud_vpc.test.uri
    subnet_uri_refs        = [arubacloud_subnet.test.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.test.uri]
  }

  settings = {
    flavor_name = "CSO4A8"
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.boot.uri
  }
}

data "arubacloud_cloudserver" "test" {
  id         = arubacloud_cloudserver.test.id
  project_id = %[1]q
}
`, projectID, osImageID)
}
