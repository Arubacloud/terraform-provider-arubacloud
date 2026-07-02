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

func TestAccSubnetResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		CheckDestroy:             testCheckSubnetDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSubnetResourceConfig(projectID, "test-subnet"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-subnet"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("vpc_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("type"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_subnet.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: ImportIDFromAttrs("arubacloud_subnet.test", "project_id", "vpc_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccSubnetResourceConfig(projectID, "test-subnet-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_subnet.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-subnet-updated"),
					),
				},
			},
		},
	})
}

func testCheckSubnetDestroyed(s *terraform.State) error {
	client, err := AccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_subnet" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		vpcID := rs.Primary.Attributes["vpc_id"]
		ref := aruba.SubnetRef(projectID, vpcID, rs.Primary.ID)
		_, err = client.Client.FromNetwork().Subnets().Get(ctx, ref)
		if provErr := provider.CheckResponseErr("get", "Subnet", err); provErr != nil {
			if provider.IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("Subnet %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSubnetResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "subnet_prereq" {
  name       = "test-acc-subnet-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "test" {
  name       = %[2]q
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.subnet_prereq.id
  type       = "Basic"

  network = {
    address = "10.0.0.0/24"
  }
}
`, projectID, name)
}
