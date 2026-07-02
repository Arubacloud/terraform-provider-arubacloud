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

func TestAccBlockStorageResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckBlockstorageDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBlockStorageResourceConfig(projectID, "test-blockstorage", 100, "Standard"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-blockstorage"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("size_gb"),
						knownvalue.Int64Exact(100),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("Standard"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_blockstorage.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_blockstorage.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccBlockStorageResourceConfig(projectID, "test-blockstorage-updated", 200, "Performance"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-blockstorage-updated"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("size_gb"),
						knownvalue.Int64Exact(200),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("Performance"),
					),
				},
			},
		},
	})
}

func TestAccBlockStorageResource_Bootable(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	osImageID := os.Getenv("ARUBACLOUD_OS_IMAGE_ID")
	if projectID == "" || osImageID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_OS_IMAGE_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckBlockstorageDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageResourceConfigBootable(projectID, osImageID, "test-bootable", 50),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-bootable"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("bootable"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_blockstorage.test",
						tfjsonpath.New("image"),
						knownvalue.StringExact(osImageID),
					),
				},
			},
		},
	})
}

func testCheckBlockstorageDestroyed(s *terraform.State) error {
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_blockstorage" {
			continue
		}
		// Use the stored URI directly rather than reconstructing the path,
		// since the SDK path format (/blockStorages/ vs /volumes/) can differ.
		uri := rs.Primary.Attributes["uri"]
		if uri == "" {
			continue
		}
		_, err = client.Client.FromStorage().Volumes().Get(ctx, aruba.URI(uri))
		if provErr := provider.CheckResponseErr("get", "Blockstorage", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("BlockStorage volume %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccBlockStorageResourceConfig(projectID, name string, sizeGB int, storageType string) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "test" {
  name           = %[2]q
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = %[3]d
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = %[4]q
}
`, projectID, name, sizeGB, storageType)
}

func testAccBlockStorageResourceConfigBootable(projectID, osImageID, name string, sizeGB int) string {
	return fmt.Sprintf(`
resource "arubacloud_blockstorage" "test" {
  name           = %[3]q
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = %[4]d
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
  bootable       = true
  image          = %[2]q
}
`, projectID, osImageID, name, sizeGB)
}
