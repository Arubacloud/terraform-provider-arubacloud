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

func TestAccVpcpeeringrouteDataSource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	vpcID := os.Getenv("ARUBACLOUD_VPC_ID")
	peeringID := os.Getenv("ARUBACLOUD_VPCPEERING_ID")
	routeID := os.Getenv("ARUBACLOUD_VPCPEERINGROUTE_ID")
	if projectID == "" || vpcID == "" || peeringID == "" || routeID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID, ARUBACLOUD_VPC_ID, ARUBACLOUD_VPCPEERING_ID and ARUBACLOUD_VPCPEERINGROUTE_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcpeeringrouteDataSourceConfig(projectID, vpcID, peeringID, routeID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("project_id"),
						knownvalue.StringExact(projectID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.StringExact(vpcID),
					),
					statecheck.ExpectKnownValue(
						"data.arubacloud_vpcpeeringroute.test",
						tfjsonpath.New("vpc_peering_id"),
						knownvalue.StringExact(peeringID),
					),
				},
			},
		},
	})
}

func testAccVpcpeeringrouteDataSourceConfig(projectID, vpcID, peeringID, routeID string) string {
	return fmt.Sprintf(`
data "arubacloud_vpcpeeringroute" "test" {
  id             = %[1]q
  project_id     = %[2]q
  vpc_id         = %[3]q
  vpc_peering_id = %[4]q
}
`, routeID, projectID, vpcID, peeringID)
}
