#!/usr/bin/env bash
# =============================================================================
# Bootstrap Script — One-time setup for Terraform remote state on Azure
# =============================================================================
# This creates:
#   1. A Resource Group for Terraform state
#   2. A Storage Account
#   3. A Blob Container for the .tfstate file
#   4. A Service Principal for GitHub Actions
#
# Run this ONCE from your local machine (requires Azure CLI logged in).
# =============================================================================

set -euo pipefail

# ---- Configuration ----
PROJECT="finops"
LOCATION="centralindia"
STATE_RG="rg-${PROJECT}-tfstate"
STATE_SA="${PROJECT}tfstate$(openssl rand -hex 2)"
STATE_CONTAINER="tfstate"

echo "=== FinOps Terraform Bootstrap ==="
echo ""
echo "Region:            $LOCATION"
echo "State RG:          $STATE_RG"
echo "State Storage:     $STATE_SA"
echo "State Container:   $STATE_CONTAINER"
echo ""

# 1. Create Resource Group for TF State
echo "→ Creating resource group for Terraform state..."
az group create \
  --name "$STATE_RG" \
  --location "$LOCATION" \
  --output none

# 2. Create Storage Account
echo "→ Creating storage account..."
az storage account create \
  --name "$STATE_SA" \
  --resource-group "$STATE_RG" \
  --location "$LOCATION" \
  --sku Standard_LRS \
  --encryption-services blob \
  --output none

# 3. Create Blob Container
echo "→ Creating blob container..."
az storage container create \
  --name "$STATE_CONTAINER" \
  --account-name "$STATE_SA" \
  --output none

# 4. Create Service Principal for GitHub Actions
echo "→ Creating service principal for CI/CD..."
SUBSCRIPTION_ID=$(az account show --query id -o tsv)

SP_OUTPUT=$(az ad sp create-for-rbac \
  --name "sp-${PROJECT}-github" \
  --role Contributor \
  --scopes "/subscriptions/${SUBSCRIPTION_ID}" \
  --sdk-auth)

echo ""
echo "=========================================="
echo "  BOOTSTRAP COMPLETE"
echo "=========================================="
echo ""
echo "Add these as GitHub Repository Secrets:"
echo ""
echo "  ARM_CLIENT_ID:         $(echo "$SP_OUTPUT" | jq -r .clientId)"
echo "  ARM_CLIENT_SECRET:     $(echo "$SP_OUTPUT" | jq -r .clientSecret)"
echo "  ARM_SUBSCRIPTION_ID:   $(echo "$SP_OUTPUT" | jq -r .subscriptionId)"
echo "  ARM_TENANT_ID:         $(echo "$SP_OUTPUT" | jq -r .tenantId)"
echo ""
echo "  TF_STATE_RESOURCE_GROUP:  $STATE_RG"
echo "  TF_STATE_STORAGE_ACCOUNT: $STATE_SA"
echo "  TF_STATE_CONTAINER:       $STATE_CONTAINER"
echo ""
echo "  PG_ADMIN_PASSWORD:     <your-postgres-password>"
echo "  JWT_SECRET:            <your-jwt-secret>"
echo "  LLM_API_KEY:           <your-llm-api-key>"
echo ""
echo "  ACR_NAME:              (from terraform output after first apply)"
echo "  AZURE_RESOURCE_GROUP:  rg-${PROJECT}-dev"
echo "  CONTAINER_APP_BACKEND: ${PROJECT}-dev-backend"
echo "  CONTAINER_APP_FRONTEND: ${PROJECT}-dev-frontend"
echo ""
echo "Storage account for state: $STATE_SA"
echo "=========================================="
