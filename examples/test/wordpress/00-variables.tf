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

variable "database_password" {
  description = "Password for the WordPress database user"
  type        = string
  sensitive   = true
  default     = "ChangeMe123!WordPress"
}

variable "wordpress_admin_password" {
  description = "Admin password for WordPress installation"
  type        = string
  sensitive   = true
  default     = "ChangeMe123!WpAdmin"
}
