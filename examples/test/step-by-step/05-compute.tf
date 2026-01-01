# Step 5: Create Compute Resources
# Note: Adjust location and zone to match your ArubaCloud region

# Keypair - SSH public key for server access
resource "arubacloud_keypair" "test" {
  name       = "test-keypair"
  location    = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  value      = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEA2No7At0tgHrcZTL0kGWyLLUqPKfOhD9hGdNV9PbJxhjOGNFxcwdQ9wCXsJ3RQaRHBuGIgVodDurrlqzxFK86yCHMgXT2YLHF0j9P4m9GDiCfOK6msbFb89p5xZExjwD2zK+w68r7iOKZeRB2yrznW5TD3KDemSPIQQIVcyLF+yxft49HWBTI3PVQ4rBVOBJ2PdC9SAOf7CYnptW24CRrC0h85szIdwMA+Kmasfl3YGzk4MxheHrTO8C40aXXpieJ9S2VQA4VJAMRyAboptIK0cKjBYrbt5YkEL0AlyBGPIu6MPYr5K/MHyDunDi9yc7VYRYRR0f46MBOSqMUiGPnMw=="
}

# Cloud Server - Virtual machine instance
# The boot_volume field accepts an image ID (like "LU22-001") which will create a bootable disk automatically
resource "arubacloud_cloudserver" "test" {
  name           = "test-cloudserver"
  location       = "ITBG-Bergamo"  # Change to your region
  project_id     = arubacloud_project.test.id
  zone           = "ITBG-1"  # Change to your zone
  vpc_id         = arubacloud_vpc.test.id
  flavor_name    = "c2.medium"  # Change to your preferred flavor
  elastic_ip_id  = arubacloud_elasticip.test.id
  boot_volume    = "LU22-001"  # Image ID - will create bootable disk automatically
  key_pair_id    = arubacloud_keypair.test.id
  subnets        = [arubacloud_subnet.test.id]
  securitygroups = [arubacloud_securitygroup.test.id]
  tags           = ["compute", "test"]
}

output "keypair_id" {
  value       = arubacloud_keypair.test.id
  description = "The ID of the created keypair"
}

output "cloudserver_id" {
  value       = arubacloud_cloudserver.test.id
  description = "The ID of the created cloud server"
}

output "cloudserver_name" {
  value       = arubacloud_cloudserver.test.name
  description = "The name of the created cloud server"
}

