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

func TestAccElasticipResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	location := os.Getenv("ARUBACLOUD_LOCATION")
	if projectID == "" || location == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_LOCATION must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckElasticipDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccElasticipResourceConfig(projectID, location, "test-elasticip"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-elasticip"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("address"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_elasticip.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_elasticip.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccElasticipResourceConfig(projectID, location, "test-elasticip-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_elasticip.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-elasticip-updated"),
					),
				},
			},
		},
	})
}

func testCheckElasticipDestroyed(s *terraform.State) error {
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_elasticip" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/network/elasticIPs/" + rs.Primary.ID)
		_, err = client.Client.FromNetwork().ElasticIPs().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Elasticip", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("ElasticIP %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccElasticipResourceConfig(projectID, location, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_elasticip" "test" {
  name       = %[3]q
  location   = %[2]q
  project_id = %[1]q
}
`, projectID, location, name)
}
