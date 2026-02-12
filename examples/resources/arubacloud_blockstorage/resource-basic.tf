# Zonal Block Storage Example
# Zonal storage is tied to a specific zone within a location
resource "arubacloud_blockstorage" "zonal_example" {
  name          = "zonal-blockstorage"
  project_id    = "your-project-id"  # Replace with your project ID
  location      = "ITBG-Bergamo"  # Change to your region
  size_gb       = 100
  billing_period = "Hour"
  zone          = "ITBG-1"  # Zonal storage - tied to this specific zone
  type          = "Standard"
  tags          = ["storage", "zonal", "test"]
}

# Regional Block Storage Example
# Regional storage is available across all zones in the location
resource "arubacloud_blockstorage" "regional_example" {
  name          = "regional-blockstorage"
  project_id    = "your-project-id"  # Replace with your project ID
  location      = "ITBG-Bergamo"  # Change to your region
  size_gb       = 100
  billing_period = "Hour"
  # zone is not specified - creates regional storage
  type          = "Standard"
  tags          = ["storage", "regional", "test"]
}
