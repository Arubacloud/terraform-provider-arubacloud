provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret
  
  log_level = "DEBUG" # Accepted: OFF, ERROR, WARN, INFO, DEBUG, TRACE. Default: OFF. Requires TF_LOG=DEBUG to surface.

  # Optional: Configure timeout for resource creation (default: 10m)
  resource_timeout = "15m"
}


