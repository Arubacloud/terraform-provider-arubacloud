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

func TestAccSnapshotResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSnapshotDestroyed,
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

func testCheckSnapshotDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_snapshot" {
			continue
		}
		resp, err := client.Client.FromStorage().Snapshots().Get(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID, nil)
		if err != nil {
			return nil
		}
		if apiErr := CheckResponse("get", "Snapshot", resp); apiErr != nil {
			if IsNotFound(apiErr) {
				continue
			}
			return apiErr
		}
		return fmt.Errorf("Snapshot %s still exists", rs.Primary.ID)
	}
	return nil
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
