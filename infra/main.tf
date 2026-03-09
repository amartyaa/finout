# -----------------------------------------------------------------------------
# Provider Configuration
# -----------------------------------------------------------------------------
provider "azurerm" {
  features {
    key_vault {
      purge_soft_delete_on_destroy = true
    }
  }
}

# -----------------------------------------------------------------------------
# Data Sources
# -----------------------------------------------------------------------------
data "azurerm_client_config" "current" {}

# Random suffix to ensure globally unique resource names
resource "random_string" "suffix" {
  length  = 4
  special = false
  upper   = false
}

# -----------------------------------------------------------------------------
# Resource Group
# -----------------------------------------------------------------------------
resource "azurerm_resource_group" "this" {
  name     = "rg-${var.project_name}-${var.environment}"
  location = var.location
  tags     = var.tags
}

# -----------------------------------------------------------------------------
# Locals
# -----------------------------------------------------------------------------
locals {
  name_prefix  = "${var.project_name}-${var.environment}"
  name_compact = "${var.project_name}${var.environment}${random_string.suffix.result}"

  default_backend_image  = "${module.acr.login_server}/finops-backend:latest"
  default_frontend_image = "${module.acr.login_server}/finops-frontend:latest"

  backend_image  = var.backend_image != "" ? var.backend_image : local.default_backend_image
  frontend_image = var.frontend_image != "" ? var.frontend_image : local.default_frontend_image
}

# -----------------------------------------------------------------------------
# Module: Container Registry
# -----------------------------------------------------------------------------
module "acr" {
  source = "./modules/container_registry"

  name                = local.name_compact
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location
  sku                 = var.acr_sku
  admin_enabled       = true
  tags                = var.tags
}

# -----------------------------------------------------------------------------
# Module: PostgreSQL Database
# -----------------------------------------------------------------------------
module "database" {
  source = "./modules/database"

  server_name         = "${local.name_prefix}-pg"
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location
  sku_name            = var.pg_sku_name
  storage_mb          = var.pg_storage_mb
  version_pg          = var.pg_version
  admin_username      = var.pg_admin_username
  admin_password      = var.pg_admin_password
  database_name       = var.pg_database_name
  tags                = var.tags
}

# -----------------------------------------------------------------------------
# Module: Redis Cache
# -----------------------------------------------------------------------------
module "redis" {
  source = "./modules/redis"

  name                = "${local.name_prefix}-redis"
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location
  sku_name            = var.redis_sku_name
  family              = var.redis_family
  capacity            = var.redis_capacity
  tags                = var.tags
}

# -----------------------------------------------------------------------------
# Module: Key Vault
# -----------------------------------------------------------------------------
module "keyvault" {
  source = "./modules/keyvault"

  name                = "${local.name_compact}kv"
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location
  tenant_id           = data.azurerm_client_config.current.tenant_id
  tags                = var.tags

  secrets = {
    "database-url" = module.database.connection_string
    "redis-url"    = module.redis.connection_string
    "jwt-secret"   = var.jwt_secret
    "llm-api-key"  = var.llm_api_key
  }
}

# -----------------------------------------------------------------------------
# Module: Container Apps (Frontend + Backend)
# -----------------------------------------------------------------------------
module "container_apps" {
  source = "./modules/container_apps"

  environment_name    = local.name_prefix
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location
  tags                = var.tags

  # ACR credentials
  acr_login_server   = module.acr.login_server
  acr_admin_username = module.acr.admin_username
  acr_admin_password = module.acr.admin_password

  # Backend config
  backend_image        = local.backend_image
  backend_cpu          = var.backend_cpu
  backend_memory       = var.backend_memory
  backend_min_replicas = var.backend_min_replicas
  backend_max_replicas = var.backend_max_replicas

  backend_env_vars = {
    SERVER_PORT  = "8080"
    LLM_BASE_URL = var.llm_base_url
    LLM_MODEL    = var.llm_model
    AWS_REGION   = "us-east-1"
  }

  backend_secret_env_vars = {
    DATABASE_URL = module.database.connection_string
    REDIS_URL    = module.redis.connection_string
    JWT_SECRET   = var.jwt_secret
    LLM_API_KEY  = var.llm_api_key
  }

  # Frontend config
  frontend_image        = local.frontend_image
  frontend_cpu          = var.frontend_cpu
  frontend_memory       = var.frontend_memory
  frontend_min_replicas = var.frontend_min_replicas
  frontend_max_replicas = var.frontend_max_replicas

  frontend_env_vars = {
    PORT = "3000"
  }

  depends_on = [module.database, module.redis, module.acr]
}
