terraform {
  required_providers {
    arubacloud = {
      source  = "hashicorp.com/arubacloud/arubacloud"
#      version = "~> 1.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret
  
  # Optional: Configure timeout for resource creation (default: 10m)
  resource_timeout = "15m"
}

variable "arubacloud_api_key" {
  description = "ArubaCloud API Key"
  type        = string
  sensitive   = true
}

variable "arubacloud_api_secret" {
  description = "ArubaCloud API Secret"
  type        = string
  sensitive   = true
}

