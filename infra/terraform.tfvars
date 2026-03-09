# -----------------------------------------------------------------------------
# Default variable values for the Dev environment
# Sensitive values (pg_admin_password, jwt_secret, llm_api_key) should be
# passed via TF_VAR_* environment variables or GitHub Secrets, never committed.
# -----------------------------------------------------------------------------

project_name = "finops"
environment  = "dev"
location     = "centralindia"

tags = {
  project     = "finops"
  environment = "dev"
  managed_by  = "terraform"
}

# PostgreSQL — Burstable B1ms for dev
pg_sku_name       = "B_Standard_B1ms"
pg_storage_mb     = 32768
pg_version        = "16"
pg_admin_username = "finopsadmin"
pg_database_name  = "finops"

# Redis — Basic C0 for dev
redis_sku_name = "Basic"
redis_family   = "C"
redis_capacity = 0

# Container Registry
acr_sku = "Basic"

# Container Apps — lightweight for dev
backend_cpu          = 0.5
backend_memory       = "1Gi"
backend_min_replicas = 0
backend_max_replicas = 3

frontend_cpu          = 0.25
frontend_memory       = "0.5Gi"
frontend_min_replicas = 0
frontend_max_replicas = 3

# LLM
llm_base_url = "https://api.openai.com/v1"
llm_model    = "gpt-4o-mini"
