# -----------------------------------------------------------------------------
# Stack Outputs
# -----------------------------------------------------------------------------

output "resource_group_name" {
  description = "Resource group name"
  value       = azurerm_resource_group.this.name
}

output "acr_login_server" {
  description = "ACR login server URL"
  value       = module.acr.login_server
}

output "acr_admin_username" {
  description = "ACR admin username"
  value       = module.acr.admin_username
}

output "postgres_fqdn" {
  description = "PostgreSQL server FQDN"
  value       = module.database.server_fqdn
}

output "redis_hostname" {
  description = "Redis cache hostname"
  value       = module.redis.hostname
}

output "keyvault_uri" {
  description = "Key Vault URI"
  value       = module.keyvault.vault_uri
}

output "backend_fqdn" {
  description = "Backend internal FQDN"
  value       = module.container_apps.backend_fqdn
}

output "frontend_url" {
  description = "Frontend public URL"
  value       = module.container_apps.frontend_url
}
