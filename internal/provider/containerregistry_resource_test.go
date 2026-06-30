package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccContainerregistryResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID must be set for acceptance tests")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckContainerregistryDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccContainerregistryResourceConfig(projectID, "test-containerregistry"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-containerregistry"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("location"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "arubacloud_containerregistry.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importIDFromAttrs("arubacloud_containerregistry.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccContainerregistryResourceConfig(projectID, "test-containerregistry-updated"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_containerregistry.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-containerregistry-updated"),
					),
				},
			},
		},
	})
}

func testCheckContainerregistryDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_containerregistry" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/providers/Aruba.Container/containerRegistries/" + rs.Primary.ID)
		_, err = client.Client.FromContainer().ContainerRegistry().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Containerregistry", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("ContainerRegistry %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccContainerregistryResourceConfig(projectID, name string) string {
	return fmt.Sprintf(`
resource "arubacloud_vpc" "cr_prereq" {
  name       = "test-acc-cr-vpc"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
}

resource "arubacloud_subnet" "cr_prereq" {
  name       = "test-acc-cr-subnet"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.cr_prereq.id
  type       = "Basic"
}

resource "arubacloud_securitygroup" "cr_prereq" {
  name       = "test-acc-cr-sg"
  location   = "ITBG-Bergamo"
  project_id = %[1]q
  vpc_id     = arubacloud_vpc.cr_prereq.id
}

resource "arubacloud_securityrule" "cr_prereq_https_ingress" {
  name              = "test-acc-cr-https-ingress"
  location          = "ITBG-Bergamo"
  project_id        = %[1]q
  vpc_id            = arubacloud_vpc.cr_prereq.id
  security_group_id = arubacloud_securitygroup.cr_prereq.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "443"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

resource "arubacloud_securityrule" "cr_prereq_egress" {
  name              = "test-acc-cr-egress"
  location          = "ITBG-Bergamo"
  project_id        = %[1]q
  vpc_id            = arubacloud_vpc.cr_prereq.id
  security_group_id = arubacloud_securitygroup.cr_prereq.id
  properties = {
    direction = "Egress"
    protocol  = "ANY"
    port      = "*"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

resource "arubacloud_elasticip" "cr_prereq" {
  name           = "test-acc-cr-eip"
  location       = "ITBG-Bergamo"
  project_id     = %[1]q
  billing_period = "Hour"
}

resource "arubacloud_blockstorage" "cr_prereq" {
  name           = "test-acc-cr-storage"
  project_id     = %[1]q
  location       = "ITBG-Bergamo"
  size_gb        = 50
  billing_period = "Hour"
  zone           = "ITBG-1"
  type           = "Standard"
}

resource "arubacloud_containerregistry" "test" {
  name           = %[2]q
  location       = "ITBG-Bergamo"
  project_id     = %[1]q
  billing_period = "Hour"

  network = {
    vpc_uri_ref            = arubacloud_vpc.cr_prereq.uri
    subnet_uri_ref         = arubacloud_subnet.cr_prereq.uri
    security_group_uri_ref = arubacloud_securitygroup.cr_prereq.uri
    public_ip_uri_ref      = arubacloud_elasticip.cr_prereq.uri
  }

  storage = {
    block_storage_uri_ref = arubacloud_blockstorage.cr_prereq.uri
  }

  settings = {
    admin_user              = "adminuser"
    concurrent_users_flavor = "Small"
  }
}
`, projectID, name)
}
