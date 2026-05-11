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

func TestAccContainerregistryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckContainerregistryDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerregistryResourceConfig("test-containerregistry"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-containerregistry"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_containerregistry.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccContainerregistryResourceConfig("test-containerregistry-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-containerregistry-updated"),
					),
				},
			},
		},
	})
}

func testCheckContainerregistryDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_containerregistry" {
			continue
		}
		resp, err := client.Client.FromContainer().ContainerRegistry().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Containerregistry", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("ContainerRegistry %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccContainerregistryResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_containerregistry" "test" {
  name           = %[1]q
  location       = "it-1"
  project_id     = "test-project-id"
  billing_period = "Hour"

  network = {
    vpc_uri_ref              = "test-vpc-uri"
    subnet_uri_ref           = "test-subnet-uri"
    security_group_uri_ref   = "test-sg-uri"
    public_ip_uri_ref        = "test-eip-uri"
  }

  storage = {
    block_storage_uri_ref = "test-storage-uri"
  }

  settings = {
    admin_user               = "admin"
    concurrent_users_flavor  = "small"
  }
}
`, name)
}
