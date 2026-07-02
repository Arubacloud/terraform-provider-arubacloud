package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used by TestUnit* tests in provider_errorcases_test.go
// which use resource.UnitTest (no TF_ACC required) and move to internal/acctest/ in the next PR.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"arubacloud": providerserver.NewProtocol6WithError(New("test")()),
}
