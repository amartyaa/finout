output "environment_id" {
  description = "Container Apps Environment ID"
  value       = azurerm_container_app_environment.this.id
}

output "backend_fqdn" {
  description = "Backend app internal FQDN"
  value       = azurerm_container_app.backend.ingress[0].fqdn
}

output "frontend_fqdn" {
  description = "Frontend app public FQDN"
  value       = azurerm_container_app.frontend.ingress[0].fqdn
}

output "frontend_url" {
  description = "Frontend public URL"
  value       = "https://${azurerm_container_app.frontend.ingress[0].fqdn}"
}
