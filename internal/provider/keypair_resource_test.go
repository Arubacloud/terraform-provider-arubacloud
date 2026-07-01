package provider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"strings"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"golang.org/x/crypto/ssh"
)

// testAccSSHPublicKey returns a valid SSH public key for acceptance tests.
// It prefers ARUBACLOUD_SSH_PUBLIC_KEY env var; falls back to a freshly generated RSA-2048 key.
func testAccSSHPublicKey(t *testing.T) string {
	t.Helper()
	if key := os.Getenv("ARUBACLOUD_SSH_PUBLIC_KEY"); key != "" {
		return strings.TrimSpace(key)
	}
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	pub, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to derive SSH public key: %v", err)
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))
}

func TestAccKeypairResource(t *testing.T) {
	projectID := os.Getenv("ARUBACLOUD_PROJECT_ID")
	location := os.Getenv("ARUBACLOUD_LOCATION")
	if projectID == "" || location == "" {
		t.Skip("ARUBACLOUD_PROJECT_ID and ARUBACLOUD_LOCATION must be set for acceptance tests")
	}
	pubKey := testAccSSHPublicKey(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testCheckKeypairDestroyed,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKeypairResourceConfig(projectID, location, "test-keypair", pubKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_keypair.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-keypair"),
					),
					statecheck.ExpectKnownValue(
						"arubacloud_keypair.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:            "arubacloud_keypair.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
				ImportStateIdFunc:       importIDFromAttrs("arubacloud_keypair.test", "project_id", "id"),
			},
			// Update and Read testing
			{
				Config: testAccKeypairResourceConfig(projectID, location, "test-keypair-updated", pubKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"arubacloud_keypair.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("test-keypair-updated"),
					),
				},
			},
		},
	})
}

func testCheckKeypairDestroyed(s *terraform.State) error {
	client, err := testAccClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "arubacloud_keypair" {
			continue
		}
		projectID := rs.Primary.Attributes["project_id"]
		ref := aruba.URI("/projects/" + projectID + "/compute/keyPairs/" + rs.Primary.ID)
		_, err = client.Client.FromCompute().KeyPairs().Get(ctx, ref)
		if provErr := CheckResponseErr("get", "Keypair", err); provErr != nil {
			if IsNotFound(provErr) {
				continue
			}
			return provErr
		}
		return fmt.Errorf("Keypair %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccKeypairResourceConfig(projectID, location, name, pubKey string) string {
	return fmt.Sprintf(`
resource "arubacloud_keypair" "test" {
  name       = %[3]q
  location   = %[2]q
  project_id = %[1]q
  value      = %[4]q
}
`, projectID, location, name, pubKey)
}
