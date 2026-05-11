package provider

// Drift-detection acceptance tests — verify that each resource's Read function
// detects out-of-band deletion (404) and removes the resource from state,
// causing Terraform to plan a recreate on the next refresh.
//
// Pattern:
//   Step 1 – Create the resource and capture its ID via a Check.
//   Step 2 – Delete it via the SDK in PreConfig, then run PlanOnly with
//             ExpectNonEmptyPlan to assert the provider detects the drift.
//
// These tests require TF_ACC=1 and a valid ARUBACLOUD_PROJECT_ID.

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVpcResource_DetectsDriftAfterOutOfBandDelete(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for drift tests")
	}

	var capturedID string
	cfg := fmt.Sprintf(`
resource "arubacloud_vpc" "drift" {
  name       = "tf-drift-vpc"
  location   = "it-1"
  project_id = %q
}
`, projectID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources["arubacloud_vpc.drift"]
					if !ok {
						return fmt.Errorf("arubacloud_vpc.drift not found in state")
					}
					capturedID = rs.Primary.ID
					return nil
				},
			},
			{
				PreConfig: func() {
					client, err := testAccClient()
					if err != nil {
						return
					}
					_, _ = client.Client.FromNetwork().VPCs().Delete(context.Background(), projectID, capturedID, nil)
					time.Sleep(15 * time.Second)
				},
				Config:             cfg,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
