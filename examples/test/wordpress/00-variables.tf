variable "arubacloud_client_id" {
  description = "ArubaCloud OAuth2 Client ID"
  type        = string
  sensitive   = true
}

variable "arubacloud_client_secret" {
  description = "ArubaCloud OAuth2 Client Secret"
  type        = string
  sensitive   = true
}

variable "database_password" {
  description = "Password for the WordPress database user"
  type        = string
  sensitive   = true
  default     = "K7m@P4z!L9"
}

variable "wordpress_admin_password" {
  description = "Admin password for WordPress installation"
  type        = string
  sensitive   = true
  default     = "ChangeMe123!WpAdmin"
}
