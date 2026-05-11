package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDbaasResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckDbaasDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbaasResourceConfig("test-dbaas"),
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
			// ImportState testing
			{
				ResourceName:      "arubacloud_dbaas.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccDbaasResourceConfig("test-dbaas-updated"),
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
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_dbaas" {
			continue
		}
		resp, err := client.Client.FromDatabase().DBaaS().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Dbaas", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("DBaaS %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDbaasResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_dbaas" "test" {
  name       = %[1]q
  location   = "it-1"
  zone       = "it-1"
  project_id = "test-project-id"
  engine_id  = "mysql-8.0"
  flavor     = "DBO2A4"

  storage = {
    size_gb = 20
  }

  network = {
    vpc_uri_ref            = "test-vpc-uri"
    subnet_uri_ref         = "test-subnet-uri"
    security_group_uri_ref = "test-sg-uri"
  }
}
`, name)
}
