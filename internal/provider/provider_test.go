package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"arubacloud": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, env := range []string{"ARUBACLOUD_CLIENT_ID", "ARUBACLOUD_CLIENT_SECRET"} {
		if os.Getenv(env) == "" {
			t.Fatalf("acceptance tests require %s to be set", env)
		}
	}
}

// importIDFromAttrs builds an ImportStateIdFunc that constructs the composite import ID
// for a resource by joining the named attributes from state with "/".
// Use "id" to include the resource's primary ID (rs.Primary.ID).
// Use any tfsdk attribute name (e.g. "project_id", "vpc_id") for other fields.
//
// Example:
//
//	ImportStateIdFunc: importIDFromAttrs("arubacloud_backup.test", "project_id", "id")
//	ImportStateIdFunc: importIDFromAttrs("arubacloud_subnet.test", "project_id", "vpc_id", "id")
func importIDFromAttrs(resourceAddr string, attrs ...string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return "", fmt.Errorf("resource %q not found in state", resourceAddr)
		}
		parts := make([]string, 0, len(attrs))
		for _, attr := range attrs {
			var v string
			if attr == "id" {
				v = rs.Primary.ID
			} else {
				v = rs.Primary.Attributes[attr]
			}
			if v == "" {
				return "", fmt.Errorf("resource %q: attribute %q is empty in state", resourceAddr, attr)
			}
			parts = append(parts, v)
		}
		return strings.Join(parts, "/"), nil
	}
}

// testAccClient builds an ArubaCloudClient from env vars for use in CheckDestroy functions.
func testAccClient() (*ArubaCloudClient, error) {
	clientID := os.Getenv("ARUBACLOUD_CLIENT_ID")
	clientSecret := os.Getenv("ARUBACLOUD_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("ARUBACLOUD_CLIENT_ID and ARUBACLOUD_CLIENT_SECRET must be set")
	}
	sdkClient, err := aruba.NewClient(aruba.DefaultOptions(clientID, clientSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to create test client: %w", err)
	}
	return &ArubaCloudClient{
		Client:          sdkClient,
		ResourceTimeout: 10 * time.Minute,
	}, nil
}
