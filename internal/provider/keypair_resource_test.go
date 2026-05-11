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

func TestAccKeypairResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckKeypairDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKeypairResourceConfig("test-keypair"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_keypair.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-keypair"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_keypair.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_keypair.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccKeypairResourceConfig("test-keypair-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_keypair.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-keypair-updated"),
					),
				},
			},
		},
	})
}

func testCheckKeypairDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_keypair" {
			continue
		}
		resp, err := client.Client.FromCompute().KeyPairs().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Keypair", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("Keypair %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccKeypairResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_keypair" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
  value      = "ssh-rsa AAAAB3NzaC1yc2EAAAA..."
}
`, name)
}
