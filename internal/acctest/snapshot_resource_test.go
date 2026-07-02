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

func TestAccSnapshotResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckSnapshotDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSnapshotResourceConfig(projectID, "test-snapshot"),
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
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_snapshot.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccSnapshotResourceConfig(projectID, "test-snapshot-updated"),
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
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_snapshot" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/snapshots/" + rs.Primary.ID)
		_, err = client.Client.FromStorage().Snapshots().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Snapshot", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("Snapshot %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSnapshotResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "snap_prereq" {
  name           = "test-acc-snap-vol"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 10
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
}

resource "arubacloud_snapshot" "test" {
  name           = %[2]q
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
  volume_uri     = arubacloud_blockstorage.snap_prereq.uri
}
`, projectID, name)
}
