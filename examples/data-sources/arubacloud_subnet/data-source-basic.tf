data "arubacloud_subnet" "basic" {
  id = "subnet-id"
}

output "subnet_name" {
  value = data.arubacloud_subnet.basic.name
}
output "subnet_location" {
  value = data.arubacloud_subnet.basic.location
}
output "subnet_project_id" {
  value = data.arubacloud_subnet.basic.project_id
}
output "subnet_vpc_id" {
  value = data.arubacloud_subnet.basic.vpc_id
}
output "subnet_address" {
  value = data.arubacloud_subnet.basic.address
}
output "subnet_dhcp_enabled" {
  value = data.arubacloud_subnet.basic.dhcp_enabled
}
output "subnet_dhcp_routes" {
  value = data.arubacloud_subnet.basic.dhcp_routes
}
