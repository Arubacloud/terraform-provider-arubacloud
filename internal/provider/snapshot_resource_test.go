package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSnapshotResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSnapshotResourceConfig("test-snapshot"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_snapshot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-snapshot"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_snapshot.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_snapshot.test",
						tfjsonpath.New("volume_uri"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_snapshot.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccSnapshotResourceConfig("test-snapshot-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_snapshot.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-snapshot-updated"),
					),
				},
			},
		},
	})
}

func testAccSnapshotResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_snapshot" "test" {
  name           = %[1]q
  project_id     = "test-project-id"
  location       = "it-1"
  billing_period = "Hour"
  volume_uri     = "/projects/test-project-id/providers/Aruba.Storage/volumes/test-volume-id"
}
`, name)
}
