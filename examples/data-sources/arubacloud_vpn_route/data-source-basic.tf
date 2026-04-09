data "arubacloud_vpnroute" "basic" {
  id             = "vpn-route-id"
  project_id     = "your-project-id"
  vpn_tunnel_id  = "your-vpn-tunnel-id"
}

output "vpnroute_name" {
  value = data.arubacloud_vpnroute.basic.name
}
output "vpnroute_destination" {
  value = data.arubacloud_vpnroute.basic.destination
}
output "vpnroute_gateway" {
  value = data.arubacloud_vpnroute.basic.gateway
}
