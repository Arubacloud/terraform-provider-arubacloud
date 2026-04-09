data "arubacloud_vpcpeering" "basic" {
  id         = "vpc-peering-id"
  project_id = "your-project-id"
  vpc_id     = "your-vpc-id"
}

output "vpcpeering_name" {
  value = data.arubacloud_vpcpeering.basic.name
}
