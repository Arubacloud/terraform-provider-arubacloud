package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSnapshotDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotDataSourceConfig(projectID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_snapshot.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_snapshot.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_snapshot.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_snapshot.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_snapshot.test",
						tfjsonpath.New("volume_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccSnapshotDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "test" {
  name           = "test-ds-blockstorage"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 10
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
}

resource "arubacloud_snapshot" "test" {
  name           = "test-ds-snapshot"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
  volume_uri     = arubacloud_blockstorage.test.uri
}

data "arubacloud_snapshot" "test" {
  id         = arubacloud_snapshot.test.id
  project_id = %[1]q
}
`, projectID)
}
