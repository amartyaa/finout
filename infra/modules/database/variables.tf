variable "server_name" {
  description = "PostgreSQL Flexible Server name"
  type        = string
}

variable "resource_group_name" {
  description = "Resource group name"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "sku_name" {
  description = "Server SKU"
  type        = string
  default     = "B_Standard_B1ms"
}

variable "storage_mb" {
  description = "Storage in MB"
  type        = number
  default     = 32768
}

variable "version_pg" {
  description = "PostgreSQL major version"
  type        = string
  default     = "16"
}

variable "admin_username" {
  description = "Admin username"
  type        = string
}

variable "admin_password" {
  description = "Admin password"
  type        = string
  sensitive   = true
}

variable "database_name" {
  description = "Database name to create"
  type        = string
  default     = "finops"
}

variable "tags" {
  description = "Tags to apply"
  type        = map(string)
  default     = {}
}
