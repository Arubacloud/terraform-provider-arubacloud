package acctest

import (
	"context"
	"crypto/rand"
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

// testAccKaasK8sVersion returns the Kubernetes version for acceptance tests.
// Falls back to ARUBACLOUD_KAAS_K8S_VERSION; uses "1.33.2" when unset.
func testAccKaasK8sVersion() string {
	if v := os.Getenv("ARUBACLOUD_KAAS_K8S_VERSION"); v != "" {
		return v
	}
	return "1.33.2"
}

func TestAccKaasResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	location := os.Getenv("ARUBACLOUD_LOCATION")
	zone := os.Getenv("ARUBACLOUD_ZONE")
	nodeInstance := os.Getenv("ARUBACLOUD_KAAS_NODE_INSTANCE")
	if projectID == "" || location == "" || zone == "" || nodeInstance == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_LOCATION, ARUBACLOUD_ZONE, and ARUBACLOUD_KAAS_NODE_INSTANCE must be set for acceptance tests")
	}
	var sfxBytes [3]byte
	rand.Read(sfxBytes[:]) //nolint:errcheck
	sfx := fmt.Sprintf("%x", sfxBytes)
	k8sVersion := testAccKaasK8sVersion()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckKaasDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKaasResourceConfig(projectID, location, zone, nodeInstance, k8sVersion, sfx, "test-kaas-"+sfx),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kaas-"+sfx),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("settings").AtMapKey("node_pools").AtSliceIndex(0).AtMapKey("zone"),
						knownvalue.NotNull(),
					),
					// Verify security_group_id is populated after provisioning (issue #333).
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("network").AtMapKey("security_group_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_kaas.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"kubeconfig",
					// The API does not return these on GET; they are only
					// known from config/prior state and cannot be recovered on import.
					"network.node_cidr.name",
					"network.security_group_name",
					"settings.node_pools",
				},
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_kaas.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccKaasResourceConfig(projectID, location, zone, nodeInstance, k8sVersion, sfx, "test-kaas-"+sfx+"-upd"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-kaas-"+sfx+"-upd"),
					),
				},
			},
		},
	})
}

func testCheckKaasDestroyed(s *terraform.State) error {
	client, err := AccClient()
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
		if provErr := provider.CheckResponseErr("get", "Kaas", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("KaaS %s still exists", rs.Primary.ID)
	}
	return nil
}

// TestAccKaas_SecurityGroupID verifies that arubacloud_kaas exposes a non-null
// network.security_group_id after provisioning, and that this ID can be used
// to look up the auto-generated security group via the arubacloud_securitygroup
// data source (end-to-end validation of the fix for issue #333).
func TestAccKaas_SecurityGroupID(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	location := os.Getenv("ARUBACLOUD_LOCATION")
	zone := os.Getenv("ARUBACLOUD_ZONE")
	nodeInstance := os.Getenv("ARUBACLOUD_KAAS_NODE_INSTANCE")
	if projectID == "" || location == "" || zone == "" || nodeInstance == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_LOCATION, ARUBACLOUD_ZONE, and ARUBACLOUD_KAAS_NODE_INSTANCE must be set for acceptance tests")
	}
	var sfxBytes [3]byte
	rand.Read(sfxBytes[:]) //nolint:errcheck
	sfx := fmt.Sprintf("%x", sfxBytes)
	k8sVersion := testAccKaasK8sVersion()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckKaasDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccKaasSecurityGroupIDConfig(projectID, location, zone, nodeInstance, k8sVersion, sfx),
				ConfigStateChecks: []statecheck.StateCheck{
					// 1. security_group_id must be a non-null, non-empty string.
					statecheck.ExpectKnownValue(
						"arubacloud_kaas.test",
						tfjsonpath.New("network").AtMapKey("security_group_id"),
						knownvalue.NotNull(),
					),
					// 2. The arubacloud_securitygroup data source lookup by that ID must succeed.
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.kaas_sg",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_securitygroup.kaas_sg",
						tfjsonpath.New("uri"),
						knownvalue.NotNull(),
					),
					// 3. The data source's security_group_id in kaas data source must also be populated.
					statecheck.ExpectKnownValue(
						"data.arubacloud_kaas.test",
						tfjsonpath.New("security_group_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccKaasSecurityGroupIDConfig(projectID, location, zone, nodeInstance, k8sVersion, sfx string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "test" {
  name       = "test-sgid-vpc-%[6]s"
  location   = %[2]q
  project_id = %[1]q
}

resource "arubacloud_subnet" "test" {
  name       = "test-sgid-subnet-%[6]s"
  location   = %[2]q
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"
}

resource "arubacloud_kaas" "test" {
  name           = "test-sgid-%[6]s"
  location       = %[2]q
  project_id     = %[1]q
  billing_period = "Hour"

  network = {
    vpc_uri_ref         = arubacloud_vpc.test.uri
    subnet_uri_ref      = arubacloud_subnet.test.uri
    security_group_name = "kaas-sgid-%[6]s"
    node_cidr = {
      address = "10.0.1.0/24"
      name    = "node-cidr-%[6]s"
    }
  }

  settings = {
    kubernetes_version = %[5]q
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

data "arubacloud_kaas" "test" {
  id         = arubacloud_kaas.test.id
  project_id = %[1]q
}

data "arubacloud_securitygroup" "kaas_sg" {
  id         = arubacloud_kaas.test.network.security_group_id
  project_id = %[1]q
}
`, projectID, location, zone, nodeInstance, k8sVersion, sfx)
}

// testAccKaasResourceConfig builds the HCL for the KaaS resource acceptance test.
// sfx is a random suffix appended to resource names to prevent same-name conflicts
// on repeated runs. KaaS creates its own security group (security_group_name); do
// NOT pre-create one with the same name or the API will reject the request.
func testAccKaasResourceConfig(projectID, location, zone, nodeInstance, k8sVersion, sfx, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "kaas_prereq" {
  name       = "test-kaas-vpc-%[6]s"
  location   = %[2]q
  project_id = %[1]q
}

resource "arubacloud_subnet" "kaas_prereq" {
  name       = "test-kaas-subnet-%[6]s"
  location   = %[2]q
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.kaas_prereq.id
  type       = "Basic"
}

resource "arubacloud_kaas" "test" {
  name           = %[7]q
  location       = %[2]q
  project_id     = %[1]q
  billing_period = "Hour"

  network = {
    vpc_uri_ref         = arubacloud_vpc.kaas_prereq.uri
    subnet_uri_ref      = arubacloud_subnet.kaas_prereq.uri
    security_group_name = "kaas-sg-%[6]s"
    node_cidr = {
      address = "10.0.1.0/24"
      name    = "node-cidr-%[6]s"
    }
  }

  settings = {
    kubernetes_version = %[5]q
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
`, projectID, location, zone, nodeInstance, k8sVersion, sfx, name)
}
