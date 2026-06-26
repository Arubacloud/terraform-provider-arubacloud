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

func TestAccVpcResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckVpcDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVpcResourceConfig("test-vpc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpc"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("project_id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_vpc.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_vpc.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccVpcResourceConfig("test-vpc-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_vpc.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-vpc-updated"),
					),
				},
			},
		},
	})
}

func testCheckVpcDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_vpc" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/network/vpcs/" + rs.Primary.ID)
		_, err = client.Client.FromNetwork().VPCs().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Vpc", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("VPC %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccVpcResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "test" {
  name       = %[1]q
  location   = "it-1"
  project_id = "test-project-id"
}
`, name)
}
