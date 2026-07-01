package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccKaasResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	location := os.Getenv("ARUBACLOUD_LOCATION")
	zone := os.Getenv("ARUBACLOUD_ZONE")
	nodeInstance := os.Getenv("ARUBACLOUD_KAAS_NODE_INSTANCE")
	if projectID == "" || location == "" || zone == "" || nodeInstance == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_LOCATION, ARUBACLOUD_ZONE, and ARUBACLOUD_KAAS_NODE_INSTANCE must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckKaasDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKaasResourceConfig(projectID, location, zone, nodeInstance, "test-kaas"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kaas"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("zone"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:            "arubacloud_kaas.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kubeconfig"},
				ImportStateIdFunc:       importIDFromAttrs("arubacloud_kaas.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccKaasResourceConfig(projectID, location, zone, nodeInstance, "test-kaas-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kaas-updated"),
					),
				},
			},
		},
	})
}

func testCheckKaasDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_kaas" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Container/kaas/" + rs.Primary.ID)
		_, err = client.Client.FromContainer().KaaS().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Kaas", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("KaaS %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccKaasResourceConfig(projectID, location, zone, nodeInstance, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "kaas_prereq" {
  name       = "test-kaas-vpc"
  location   = %[2]q
  project_id = %[1]q
}

resource "arubacloud_subnet" "kaas_prereq" {
  name       = "test-kaas-subnet"
  location   = %[2]q
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.kaas_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "kaas_prereq" {
  name       = "test-kaas-sg"
  location   = %[2]q
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.kaas_prereq.id
}

resource "arubacloud_kaas" "test" {
  name           = %[5]q
  location       = %[2]q
  project_id     = %[1]q
  billing_period = "Hour"

  network = {
    vpc_uri_ref         = arubacloud_vpc.kaas_prereq.uri
    subnet_uri_ref      = arubacloud_subnet.kaas_prereq.uri
    security_group_name = arubacloud_securitygroup.kaas_prereq.name
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
        instance = %[4]q
        zone     = %[3]q
      }
    ]
  }
}
`, projectID, location, zone, nodeInstance, name)
}
