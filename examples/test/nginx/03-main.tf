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