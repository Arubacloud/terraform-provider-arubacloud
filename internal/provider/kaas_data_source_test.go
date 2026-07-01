package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccKaasDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	zone := os.Getenv("ARUBACLOUD_ZONE")
	nodeInstance := os.Getenv("ARUBACLOUD_KAAS_NODE_INSTANCE")
	if projectID == "" || zone == "" || nodeInstance == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_ZONE, and ARUBACLOUD_KAAS_NODE_INSTANCE must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKaasDataSourceConfig(projectID, zone, nodeInstance),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_kaas.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_kaas.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_kaas.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
				},
			},
		},
	})
}

func testAccKaasDataSourceConfig(projectID, zone, nodeInstance string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "test" {
  name       = "test-ds-kaas-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "test" {
  name       = "test-ds-kaas-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "test" {
  name       = "test-ds-kaas-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
}

resource "arubacloud_kaas" "test" {
  name           = "test-ds-kaas"
  location       = "ITBG-Bergamo"
  project_id     = %[1]q
  billing_period = "Hour"

  network = {
    vpc_uri_ref         = arubacloud_vpc.test.uri
    subnet_uri_ref      = arubacloud_subnet.test.uri
    security_group_name = arubacloud_securitygroup.test.name
    node_cidr = {
      address = "10.0.1.0/24"
      name    = "node-cidr"
    }
  }

  settings = {
    kubernetes_version = "1.28"
    ha                 = false
    node_pools = [
      {
        name     = "default"
        nodes    = 1
        instance = %[3]q
        zone     = %[2]q
      }
    ]
  }
}

data "arubacloud_kaas" "test" {
  id         = arubacloud_kaas.test.id
  project_id = %[1]q
}
`, projectID, zone, nodeInstance)
}
