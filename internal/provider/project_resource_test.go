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

func TestAccProjectResource(t *testing.T) {
	t.Skip("Skipping until project update API validation bug is fixed")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckProjectDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig("test-project"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_project.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-project"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_project.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccProjectResourceConfig("test-project-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_project.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-project-updated"),
					),
				},
			},
		},
	})
}

func testCheckProjectDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_project" {
			continue
		}
		resp, err := client.Client.FromProject().Get(ctx, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if apiErr := CheckResponse("get", "Project", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("Project %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccProjectResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_project" "test" {
  name = %[1]q
}
`, name)
}
