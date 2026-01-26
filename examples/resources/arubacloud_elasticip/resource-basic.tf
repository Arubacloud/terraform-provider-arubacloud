resource "arubacloud_elasticip" "example" {
  name           = "example-elastic-ip"
  location       = "ITBG-Bergamo"  # Change to your region
  project_id     = "your-project-id"  # Replace with your project ID
  billing_period = "hourly"  # Required: "hourly", "monthly", or "yearly"
  tags           = ["public", "test"]
}

# Output the Elastic IP address (computed field from ElasticIpPropertiesResponse)
output "elastic_ip_address" {
  value       = arubacloud_elasticip.example.address
  description = "The IP address of the created Elastic IP (computed from ElasticIpPropertiesResponse)"
}
