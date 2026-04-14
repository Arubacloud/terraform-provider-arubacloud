resource "arubacloud_vpnroute" "example" {
  name          = "example-vpn-route"
  location      = "ITBG-Bergamo"
  project_id    = "your-project-id"
  vpn_tunnel_id = "your-vpn-tunnel-id"

  properties = {
    cloud_subnet  = "10.0.1.0/24"
    on_prem_subnet = "192.168.1.0/24"
  }
}
