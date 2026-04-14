resource "arubacloud_vpntunnel" "example" {
  name       = "example-vpn-tunnel"
  location   = "ITBG-Bergamo"
  project_id = "your-project-id"

  properties = {
    vpn_type            = "Site-To-Site"
    vpn_client_protocol = "ikev2"
  }
}
