variable "environment_name" {
  description = "Container Apps Environment name"
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

# --- ACR ---
variable "acr_login_server" {
  description = "ACR login server URL"
  type        = string
}

variable "acr_admin_username" {
  description = "ACR admin username"
  type        = string
}

variable "acr_admin_password" {
  description = "ACR admin password"
  type        = string
  sensitive   = true
}

# --- Backend ---
variable "backend_image" {
  description = "Backend container image"
  type        = string
}

variable "backend_cpu" {
  description = "Backend CPU"
  type        = number
  default     = 0.5
}

variable "backend_memory" {
  description = "Backend memory"
  type        = string
  default     = "1Gi"
}

variable "backend_min_replicas" {
  description = "Backend min replicas"
  type        = number
  default     = 0
}

variable "backend_max_replicas" {
  description = "Backend max replicas"
  type        = number
  default     = 3
}

variable "backend_env_vars" {
  description = "Backend environment variables (non-secret)"
  type        = map(string)
  default     = {}
}

variable "backend_secret_env_vars" {
  description = "Backend secret environment variables"
  type        = map(string)
  default     = {}
  sensitive   = true
}

# --- Frontend ---
variable "frontend_image" {
  description = "Frontend container image"
  type        = string
}

variable "frontend_cpu" {
  description = "Frontend CPU"
  type        = number
  default     = 0.25
}

variable "frontend_memory" {
  description = "Frontend memory"
  type        = string
  default     = "0.5Gi"
}

variable "frontend_min_replicas" {
  description = "Frontend min replicas"
  type        = number
  default     = 0
}

variable "frontend_max_replicas" {
  description = "Frontend max replicas"
  type        = number
  default     = 3
}

variable "frontend_env_vars" {
  description = "Frontend environment variables (non-secret)"
  type        = map(string)
  default     = {}
}

variable "tags" {
  description = "Tags to apply"
  type        = map(string)
  default     = {}
}
