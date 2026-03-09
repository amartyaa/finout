# -----------------------------------------------------------------------------
# Log Analytics Workspace (required by Container Apps Environment)
# -----------------------------------------------------------------------------
resource "azurerm_log_analytics_workspace" "this" {
  name                = "${var.environment_name}-logs"
  resource_group_name = var.resource_group_name
  location            = var.location
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = var.tags
}

# -----------------------------------------------------------------------------
# Container Apps Environment
# -----------------------------------------------------------------------------
resource "azurerm_container_app_environment" "this" {
  name                       = var.environment_name
  resource_group_name        = var.resource_group_name
  location                   = var.location
  log_analytics_workspace_id = azurerm_log_analytics_workspace.this.id
  tags                       = var.tags
}

# -----------------------------------------------------------------------------
# Backend Container App (internal ingress)
# -----------------------------------------------------------------------------
locals {
  # Convert secret env vars to the secrets block format
  backend_secrets = [
    for key, value in var.backend_secret_env_vars : {
      name  = lower(replace(key, "_", "-"))
      value = value
    }
  ]

  # Map env var names to secret references
  backend_secret_env = [
    for key, value in var.backend_secret_env_vars : {
      name        = key
      secret_name = lower(replace(key, "_", "-"))
    }
  ]

  # Plain env vars
  backend_plain_env = [
    for key, value in var.backend_env_vars : {
      name  = key
      value = value
    }
  ]
}

resource "azurerm_container_app" "backend" {
  name                         = "${var.environment_name}-backend"
  resource_group_name          = var.resource_group_name
  container_app_environment_id = azurerm_container_app_environment.this.id
  revision_mode                = "Single"
  tags                         = var.tags

  # ACR credentials
  registry {
    server               = var.acr_login_server
    username             = var.acr_admin_username
    password_secret_name = "acr-password"
  }

  # Secrets block — ACR password + app secrets
  secret {
    name  = "acr-password"
    value = var.acr_admin_password
  }

  dynamic "secret" {
    for_each = { for secret in local.backend_secrets : secret.name => secret }
    content {
      name  = secret.value.name
      value = secret.value.value
    }
  }

  template {
    min_replicas = var.backend_min_replicas
    max_replicas = var.backend_max_replicas

    container {
      name   = "backend"
      image  = var.backend_image
      cpu    = var.backend_cpu
      memory = var.backend_memory

      # Plain environment variables
      dynamic "env" {
        for_each = local.backend_plain_env
        content {
          name  = env.value.name
          value = env.value.value
        }
      }

      # Secret environment variables
      dynamic "env" {
        for_each = { for env in local.backend_secret_env : env.name => env }
        content {
          name        = env.value.name
          secret_name = env.value.secret_name
        }
      }

      liveness_probe {
        transport = "HTTP"
        path      = "/health"
        port      = 8080
      }

      readiness_probe {
        transport = "HTTP"
        path      = "/health"
        port      = 8080
      }
    }
  }

  ingress {
    external_enabled = false
    target_port      = 8080
    transport        = "http"

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }
}

# -----------------------------------------------------------------------------
# Frontend Container App (external ingress)
# -----------------------------------------------------------------------------
resource "azurerm_container_app" "frontend" {
  name                         = "${var.environment_name}-frontend"
  resource_group_name          = var.resource_group_name
  container_app_environment_id = azurerm_container_app_environment.this.id
  revision_mode                = "Single"
  tags                         = var.tags

  # ACR credentials
  registry {
    server               = var.acr_login_server
    username             = var.acr_admin_username
    password_secret_name = "acr-password"
  }

  secret {
    name  = "acr-password"
    value = var.acr_admin_password
  }

  template {
    min_replicas = var.frontend_min_replicas
    max_replicas = var.frontend_max_replicas

    container {
      name   = "frontend"
      image  = var.frontend_image
      cpu    = var.frontend_cpu
      memory = var.frontend_memory

      # Pass the backend internal URL to the frontend
      env {
        name  = "NEXT_PUBLIC_API_URL"
        value = "https://${azurerm_container_app.backend.ingress[0].fqdn}"
      }

      dynamic "env" {
        for_each = var.frontend_env_vars
        content {
          name  = env.key
          value = env.value
        }
      }
    }
  }

  ingress {
    external_enabled = true
    target_port      = 3000
    transport        = "http"

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }
}
