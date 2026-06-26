package provider

import (
	"context"
	"fmt"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccElasticipResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckElasticipDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccElasticipResourceConfig("test-elasticip"),
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
				ImportStateIdFunc: importIDFromAttrs("arubacloud_elasticip.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccElasticipResourceConfig("test-elasticip-updated"),
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
	client, err := testAccClient()
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
		if provErr := CheckResponseErr("get", "Elasticip", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("ElasticIP %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccElasticipResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_elasticip" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
}
`, name)
}
