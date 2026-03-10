terraform {
  backend "azurerm" {
    resource_group_name  = "rg-finops-tfstate"
    storage_account_name = "finopstfstatef3af"
    container_name       = "tfstate"
    key                  = "finops.terraform.tfstate"
  }
}
