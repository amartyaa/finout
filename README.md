# AI Cloud FinOps Optimizer

A cloud cost management Minimum Viable Product (MVP) designed to ingest, process, and optimize cloud billing data using AI. The system provides multi-tenant cost tracking, anomaly detection, spend forecasting, and savings recommendations.

## Core Components

- **Frontend**: Next.js (TypeScript) dashboard with a dark glassmorphism design for visualizing cost KPIs, AI insights, and managing cloud connections.
- **Backend API**: Go (Gin framework) RESTful service handling authentication, multi-tenant RBAC, cloud integrations, and core business logic.
- **Background Workers**: Go routines listening on Redis queues to periodically sync cost data asynchronously.
- **Database**: PostgreSQL storing user accounts, organization hierarchies, cloud credentials, normalized time-series cost data, and cached AI insights.
- **AI Engine**: A pipeline that combines statistical calculations (e.g., rolling means, linear regression) with Large Language Models (LLMs) to generate human-readable narratives, forecast analyses, and actionable optimizations.

## Architecture & Data Flow

1. **Multi-Tenancy**: Users belong to Organizations. All queries and background jobs are hard-scoped to the organization (tenant boundary).
2. **Cloud Integration**: 
   - **AWS**: Connects via Cross-Account IAM Role Assumption and queries the Cost Explorer API.
   - **Azure**: Connects via Service Principal OAuth2 and queries the Cost Management API.
3. **Data Ingestion**: Redis job queues trigger background workers to fetch, normalize, and upsert daily service-level costs into a unified `daily_costs` PostgreSQL table.
4. **AI Generation**: Post-sync, the worker analyzes the daily costs to detect anomalies (cost spikes), project month-end forecasts, and identify idle/underutilized resources.
5. **Insights**: The Next.js dashboard presents the AI-generated narratives alongside standard time-series visualizations.
