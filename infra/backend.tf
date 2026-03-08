# ──────────────────────────────────────────────
# Terraform Remote State in Azure Storage
# ──────────────────────────────────────────────
# BOOTSTRAP (one-time manual steps):
#
#   az group create --name rg-finops-tfstate --location centralindia
#
#   az storage account create \
#     --name finoutfinopstfstate \
#     --resource-group rg-finops-tfstate \
#     --location centralindia \
#     --sku Standard_LRS \
#     --encryption-services blob
#
#   az storage container create \
#     --name tfstate \
#     --account-name finoutfinopstfstate
# ──────────────────────────────────────────────

terraform {
  backend "azurerm" {
    resource_group_name  = "rg-finops-tfstate"
    storage_account_name = "finoutfinopstfstate"
    container_name       = "tfstate"
    key                  = "finops.terraform.tfstate"
  }
}
