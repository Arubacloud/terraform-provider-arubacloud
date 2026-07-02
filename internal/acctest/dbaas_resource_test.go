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

func TestAccDbaasResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckDbaasDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDbaasResourceConfig(projectID, "test-dbaas"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_dbaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-dbaas"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_dbaas.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_dbaas.test",
						tfjsonpath.New("engine_id"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				ResourceName:      "arubacloud_dbaas.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_dbaas.test", "project_id", "id"),
				// zone, engine_id and network.* are not returned by the API GET response.
				// Read preserves them from prior state, but on import the initial state
				// only has project_id + id so they remain null after the first Read.
				// Users must re-specify these values in their config after import.
				ImportStateVerifyIgnore: []string{
					"zone",
					"engine_id",
					"network.%",
					"network.vpc_uri_ref",
					"network.subnet_uri_ref",
					"network.security_group_uri_ref",
					"network.elastic_ip_uri_ref",
				},
			},
			{
				Config: testAccDbaasResourceConfig(projectID, "test-dbaas-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_dbaas.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-dbaas-updated"),
					),
				},
			},
		},
	})
}

func testCheckDbaasDestroyed(s *terraform.State) error {
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_dbaas" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + rs.Primary.ID)
		_, err = client.Client.FromDatabase().DBaaS().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Dbaas", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("DBaaS %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDbaasResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "dbaas_prereq" {
  name       = "test-acc-dbaas-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "dbaas_prereq" {
  name       = "test-acc-dbaas-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.dbaas_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "dbaas_prereq" {
  name       = "test-acc-dbaas-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.dbaas_prereq.id
}

resource "arubacloud_dbaas" "test" {
  name       = %[2]q
  location   = "ITBG-Bergamo"
  zone       = "ITBG-1"
  project_id = %[1]q
  engine_id  = "mysql-8.0"
  flavor     = "DBO2A4"

  storage = {
    size_gb = 20
  }

  network = {
    vpc_uri_ref            = arubacloud_vpc.dbaas_prereq.uri
    subnet_uri_ref         = arubacloud_subnet.dbaas_prereq.uri
    security_group_uri_ref = arubacloud_securitygroup.dbaas_prereq.uri
  }
}
`, projectID, name)
}
