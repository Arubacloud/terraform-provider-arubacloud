data "arubacloud_vpcpeeringroute" "basic" {
  id             = "route-name"
  project_id     = "your-project-id"
  vpc_id         = "your-vpc-id"
  vpc_peering_id = "your-vpc-peering-id"
}

output "vpcpeeringroute_name" {
  value = data.arubacloud_vpcpeeringroute.basic.name
}
