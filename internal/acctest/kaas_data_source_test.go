package acctest

import (
	"crypto/rand"
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

	var sfxBytes [3]byte
	rand.Read(sfxBytes[:]) //nolint:errcheck
	sfx := fmt.Sprintf("%x", sfxBytes)
	k8sVersion := testAccKaasK8sVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKaasDataSourceConfig(projectID, zone, nodeInstance, k8sVersion, sfx),
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

// testAccKaasDataSourceConfig builds the HCL for the KaaS data source acceptance test.
// KaaS creates its own security group (security_group_name); do NOT pre-create one.
func testAccKaasDataSourceConfig(projectID, zone, nodeInstance, k8sVersion, sfx string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "test" {
  name       = "test-ds-kaas-vpc-%[5]s"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "test" {
  name       = "test-ds-kaas-subnet-%[5]s"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"
}

resource "arubacloud_kaas" "test" {
  name           = "test-ds-kaas-%[5]s"
  location       = "ITBG-Bergamo"
  project_id     = %[1]q
  billing_period = "Hour"

  network = {
    vpc_uri_ref         = arubacloud_vpc.test.uri
    subnet_uri_ref      = arubacloud_subnet.test.uri
    security_group_name = "kaas-ds-sg-%[5]s"
    node_cidr = {
      address = "10.0.1.0/24"
      name    = "node-cidr-%[5]s"
    }
  }

  settings = {
    kubernetes_version = %[4]q
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
`, projectID, zone, nodeInstance, k8sVersion, sfx)
}
