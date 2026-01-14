output "elastic_ip_address" {
  value       = arubacloud_elasticip.test.address
  description = "The public IP address of the Elastic IP (may be null until assigned)"
}

output "nginx_test_command" {
  value       = arubacloud_elasticip.test.address != null ? format("curl http://%s:80", arubacloud_elasticip.test.address) : "curl http://<elastic-ip-address>:80 (address will be available after apply)"
  description = "Command to test nginx on the cloud server"
}

output "nginx_url" {
  value       = arubacloud_elasticip.test.address != null ? format("http://%s", arubacloud_elasticip.test.address) : "http://<elastic-ip-address> (address will be available after apply)"
  description = "URL to access nginx on the cloud server"
}