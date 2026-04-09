data "arubacloud_vpntunnel" "basic" {
  id         = "vpn-tunnel-id"
  project_id = "your-project-id"
}

output "vpntunnel_name" {
  value = data.arubacloud_vpntunnel.basic.name
}
