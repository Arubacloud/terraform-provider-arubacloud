package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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
