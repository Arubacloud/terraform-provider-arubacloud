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

func TestAccKaasResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckKaasDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKaasResourceConfig("test-kaas"),
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
				ResourceName:      "arubacloud_kaas.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccKaasResourceConfig("test-kaas-updated"),
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
		resp, err := client.Client.FromContainer().KaaS().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Kaas", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("KaaS %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccKaasResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_kaas" "test" {
  name           = %[1]q
  location       = "it-1"
  project_id     = "test-project-id"
  billing_period = "Hour"

  network = {
    vpc_uri_ref         = "test-vpc-uri"
    subnet_uri_ref      = "test-subnet-uri"
    security_group_name = "test-sg"
    node_cidr = {
      address = "10.0.1.0/24"
      name    = "node-cidr"
    }
  }

  settings = {
    kubernetes_version = "1.28"
    ha                 = false
    node_pools = []
  }
}
`, name)
}
