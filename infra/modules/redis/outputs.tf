output "id" {
  description = "Redis Cache resource ID"
  value       = azurerm_redis_cache.this.id
}

output "hostname" {
  description = "Redis hostname"
  value       = azurerm_redis_cache.this.hostname
}

output "port" {
  description = "Redis SSL port"
  value       = azurerm_redis_cache.this.ssl_port
}

output "primary_access_key" {
  description = "Redis primary access key"
  value       = azurerm_redis_cache.this.primary_access_key
  sensitive   = true
}

output "connection_string" {
  description = "Redis connection string (TLS)"
  value       = "rediss://:${azurerm_redis_cache.this.primary_access_key}@${azurerm_redis_cache.this.hostname}:${azurerm_redis_cache.this.ssl_port}/0"
  sensitive   = true
}
