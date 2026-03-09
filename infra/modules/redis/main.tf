resource "azurerm_redis_cache" "this" {
  name                = var.name
  resource_group_name = var.resource_group_name
  location            = var.location
  capacity            = var.capacity
  family              = var.family
  sku_name            = var.sku_name
  minimum_tls_version = "1.2"
  enable_non_ssl_port = false

  redis_configuration {}

  tags = var.tags
}
