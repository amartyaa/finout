terraform {
  backend "azurerm" {
    resource_group_name  = "rg-finops-tfstate"
    storage_account_name = "finopstfstate"
    container_name       = "tfstate"
    key                  = "finops.terraform.tfstate"
  }
}
