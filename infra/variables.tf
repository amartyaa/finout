# -----------------------------------------------------------------------------
# General
# -----------------------------------------------------------------------------
variable "project_name" {
  description = "Project name used as a prefix for all resources"
  type        = string
  default     = "finops"
}

variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "location" {
  description = "Azure region for all resources"
  type        = string
  default     = "centralindia"
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# -----------------------------------------------------------------------------
# PostgreSQL
# -----------------------------------------------------------------------------
variable "pg_sku_name" {
  description = "PostgreSQL Flexible Server SKU (e.g. B_Standard_B1ms, GP_Standard_D2s_v3)"
  type        = string
  default     = "B_Standard_B1ms"
}

variable "pg_storage_mb" {
  description = "PostgreSQL storage in MB"
  type        = number
  default     = 32768
}

variable "pg_version" {
  description = "PostgreSQL major version"
  type        = string
  default     = "16"
}

variable "pg_admin_username" {
  description = "PostgreSQL admin username"
  type        = string
  default     = "finopsadmin"
}

variable "pg_admin_password" {
  description = "PostgreSQL admin password"
  type        = string
  sensitive   = true
}

variable "pg_database_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "finops"
}

# -----------------------------------------------------------------------------
# Redis
# -----------------------------------------------------------------------------
variable "redis_sku_name" {
  description = "Redis Cache SKU (Basic, Standard, Premium)"
  type        = string
  default     = "Basic"
}

variable "redis_family" {
  description = "Redis Cache family (C for Basic/Standard, P for Premium)"
  type        = string
  default     = "C"
}

variable "redis_capacity" {
  description = "Redis Cache capacity (0-6 for C family, 1-5 for P family)"
  type        = number
  default     = 0
}

# -----------------------------------------------------------------------------
# Container Registry
# -----------------------------------------------------------------------------
variable "acr_sku" {
  description = "Azure Container Registry SKU (Basic, Standard, Premium)"
  type        = string
  default     = "Basic"
}

# -----------------------------------------------------------------------------
# Container Apps
# -----------------------------------------------------------------------------
variable "backend_image" {
  description = "Backend container image (fully qualified)"
  type        = string
  default     = ""
}

variable "frontend_image" {
  description = "Frontend container image (fully qualified)"
  type        = string
  default     = ""
}

variable "backend_cpu" {
  description = "Backend container CPU cores"
  type        = number
  default     = 0.5
}

variable "backend_memory" {
  description = "Backend container memory (e.g. 1Gi)"
  type        = string
  default     = "1Gi"
}

variable "frontend_cpu" {
  description = "Frontend container CPU cores"
  type        = number
  default     = 0.25
}

variable "frontend_memory" {
  description = "Frontend container memory (e.g. 0.5Gi)"
  type        = string
  default     = "0.5Gi"
}

variable "backend_min_replicas" {
  description = "Backend minimum replica count"
  type        = number
  default     = 0
}

variable "backend_max_replicas" {
  description = "Backend maximum replica count"
  type        = number
  default     = 3
}

variable "frontend_min_replicas" {
  description = "Frontend minimum replica count"
  type        = number
  default     = 0
}

variable "frontend_max_replicas" {
  description = "Frontend maximum replica count"
  type        = number
  default     = 3
}

# -----------------------------------------------------------------------------
# App Secrets
# -----------------------------------------------------------------------------
variable "jwt_secret" {
  description = "JWT signing secret"
  type        = string
  sensitive   = true
}

variable "llm_api_key" {
  description = "LLM provider API key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "llm_base_url" {
  description = "LLM provider base URL"
  type        = string
  default     = "https://api.openai.com/v1"
}

variable "llm_model" {
  description = "LLM model name"
  type        = string
  default     = "gpt-4o-mini"
}
