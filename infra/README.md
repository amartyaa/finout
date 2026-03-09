# FinOps — Azure Infrastructure

Terraform-managed infrastructure for the FinOps application, deployed to **Azure Central India (Pune)**.

## Architecture

| Component | Azure Service | Config |
|---|---|---|
| Frontend | Container Apps (external) | Next.js on port 3000 |
| Backend | Container Apps (internal) | Go on port 8080 |
| Database | PostgreSQL Flexible Server | v16, Burstable B1ms |
| Cache | Azure Cache for Redis | Basic C0 |
| Registry | Azure Container Registry | Basic |
| Secrets | Azure Key Vault | Standard |

## Quick Start

### 1. Bootstrap (one-time)

```bash
# Login to Azure CLI
az login

# Run the bootstrap script
chmod +x infra/scripts/bootstrap.sh
./infra/scripts/bootstrap.sh
```

This creates the remote state storage and outputs all the secrets you need.

### 2. Configure GitHub Secrets

Add all the secrets from the bootstrap output to your GitHub repository:
- `Settings` → `Secrets and variables` → `Actions` → `New repository secret`

Required secrets:
| Secret | Source |
|---|---|
| `ARM_CLIENT_ID` | Bootstrap output |
| `ARM_CLIENT_SECRET` | Bootstrap output |
| `ARM_SUBSCRIPTION_ID` | Bootstrap output |
| `ARM_TENANT_ID` | Bootstrap output |
| `TF_STATE_RESOURCE_GROUP` | Bootstrap output |
| `TF_STATE_STORAGE_ACCOUNT` | Bootstrap output |
| `TF_STATE_CONTAINER` | Bootstrap output |
| `PG_ADMIN_PASSWORD` | Your choice (strong password) |
| `JWT_SECRET` | Your choice (random string) |
| `LLM_API_KEY` | Your OpenAI/LLM API key |
| `ACR_NAME` | From Terraform output (after first apply) |
| `AZURE_RESOURCE_GROUP` | `rg-finops-dev` |
| `CONTAINER_APP_BACKEND` | `finops-dev-backend` |
| `CONTAINER_APP_FRONTEND` | `finops-dev-frontend` |

### 3. (Optional) Create a GitHub Environment

For extra safety, create a `production` environment in GitHub:
- `Settings` → `Environments` → `New environment` → `production`
- Add required reviewers so infra changes need approval before apply

### 4. Deploy

Push changes to the `infra/` folder:

```bash
git add infra/
git commit -m "feat: add Azure infrastructure"
git push origin main
```

The `Infra: Apply` workflow will run automatically.

## Workflows

| Workflow | Trigger | What it does |
|---|---|---|
| `infra-plan.yml` | PR touching `infra/**` | Validates and plans, posts to PR |
| `infra-apply.yml` | Push to main touching `infra/**` | Applies infrastructure changes |
| `buildall.yml` | Push to main touching `frontend/**` or `backend/**` | Builds images, pushes to ACR, deploys to Container Apps |

## Cost (Dev — Central India)

| Service | ~INR/month |
|---|---|
| Container Apps (2 apps, consumption) | ₹0–1,260 |
| PostgreSQL Flexible Server (B1ms) | ₹1,260–2,100 |
| Azure Cache for Redis (Basic C0) | ₹1,344 |
| Azure Container Registry (Basic) | ₹420 |
| Key Vault | ₹42 |
| **Total** | **₹3,066–5,166** |

Under the limit of Free Azure Account 😉

## File Structure

```
infra/
├── main.tf                 # Root — wires all modules
├── variables.tf            # Input variables
├── outputs.tf              # Stack outputs
├── terraform.tfvars        # Default values (non-secret)
├── versions.tf             # Provider versions
├── backend.tf              # Remote state config
├── .gitignore              # Ignore state/lock files
├── scripts/
│   └── bootstrap.sh        # One-time Azure setup
└── modules/
    ├── container_registry/  # ACR
    ├── database/            # PostgreSQL Flexible Server
    ├── redis/               # Azure Cache for Redis
    ├── container_apps/      # Frontend + Backend apps
    └── keyvault/            # Secret storage
```
