package provider

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
				ImportStateIdFunc: importIDFromAttrs("arubacloud_subnet.test", "project_id", "vpc_id", "id"),
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
	client, err := testAccClient()
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
		if provErr := CheckResponseErr("get", "Subnet", err); provErr != nil {
			if IsNotFound(provErr) {
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

// ── cidrContains ──────────────────────────────────────────────────────────────

func TestCidrContains(t *testing.T) {
	parseCIDR := func(s string) *net.IPNet {
		t.Helper()
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatalf("ParseCIDR(%q): %v", s, err)
		}
		return n
	}

	cases := []struct {
		parent string
		child  string
		want   bool
	}{
		// Subnet is contained in the parent.
		{"10.0.0.0/8", "10.1.2.0/24", true},
		// Equal networks are considered contained.
		{"10.0.0.0/24", "10.0.0.0/24", true},
		// Child prefix longer than parent — still a subset.
		{"192.168.0.0/16", "192.168.1.0/24", true},
		// Child has a different network address — not contained.
		{"10.0.0.0/24", "10.0.1.0/24", false},
		// Parent is a subnet of child (narrower mask) — not contained.
		{"10.0.0.0/24", "10.0.0.0/16", false},
		// IPv4 vs IPv6 — different bit widths, never contained.
		{"10.0.0.0/8", "::1/128", false},
	}

	for _, tc := range cases {
		parent, child := parseCIDR(tc.parent), parseCIDR(tc.child)
		if got := cidrContains(parent, child); got != tc.want {
			t.Errorf("cidrContains(%q, %q) = %v, want %v", tc.parent, tc.child, got, tc.want)
		}
	}
}
