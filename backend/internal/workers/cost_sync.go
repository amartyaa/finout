package workers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/finops/backend/internal/ai"
	awsclient "github.com/finops/backend/internal/aws"
	azureclient "github.com/finops/backend/internal/azure"
	"github.com/finops/backend/internal/config"
)

type CostSyncWorker struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Config *config.Config
}

func NewCostSyncWorker(db *pgxpool.Pool, rdb *redis.Client, cfg *config.Config) *CostSyncWorker {
	return &CostSyncWorker{DB: db, Redis: rdb, Config: cfg}
}

// Start listens for AWS cost sync jobs
func (w *CostSyncWorker) Start(ctx context.Context) {
	log.Println("🔄 AWS cost sync worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("AWS cost sync worker shutting down")
			return
		default:
			result, err := w.Redis.BLPop(ctx, 5*time.Second, "jobs:cost_sync").Result()
			if err != nil {
				continue
			}

			if len(result) < 2 {
				continue
			}

			orgID := result[1]
			log.Printf("📊 Starting AWS cost sync for org: %s", orgID)

			if err := w.syncAWSCosts(ctx, orgID); err != nil {
				log.Printf("❌ AWS cost sync failed for org %s: %v", orgID, err)
				w.DB.Exec(ctx,
					"UPDATE aws_connections SET status = 'error', error_message = $2, updated_at = NOW() WHERE org_id = $1",
					orgID, err.Error(),
				)
			} else {
				log.Printf("✅ AWS cost sync completed for org: %s", orgID)
			}
		}
	}
}

// StartAzure listens for Azure cost sync jobs
func (w *CostSyncWorker) StartAzure(ctx context.Context) {
	log.Println("🔄 Azure cost sync worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Azure cost sync worker shutting down")
			return
		default:
			result, err := w.Redis.BLPop(ctx, 5*time.Second, "jobs:azure_cost_sync").Result()
			if err != nil {
				continue
			}

			if len(result) < 2 {
				continue
			}

			orgID := result[1]
			log.Printf("📊 Starting Azure cost sync for org: %s", orgID)

			if err := w.syncAzureCosts(ctx, orgID); err != nil {
				log.Printf("❌ Azure cost sync failed for org %s: %v", orgID, err)
				w.DB.Exec(ctx,
					"UPDATE azure_connections SET status = 'error', error_message = $2, updated_at = NOW() WHERE org_id = $1",
					orgID, err.Error(),
				)
			} else {
				log.Printf("✅ Azure cost sync completed for org: %s", orgID)
			}
		}
	}
}

func (w *CostSyncWorker) syncAWSCosts(ctx context.Context, orgID string) error {
	var roleARN, externalID string
	err := w.DB.QueryRow(ctx,
		"SELECT role_arn, external_id FROM aws_connections WHERE org_id = $1",
		orgID,
	).Scan(&roleARN, &externalID)
	if err != nil {
		return fmt.Errorf("no AWS connection found: %w", err)
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	ceClient := awsclient.NewCostExplorerClient(w.Config)
	costs, err := ceClient.FetchDailyCosts(ctx, roleARN, externalID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch AWS costs: %w", err)
	}

	for _, cost := range costs {
		_, err := w.DB.Exec(ctx,
			`INSERT INTO daily_costs (org_id, date, service, account_id, amount, currency, cloud_provider)
			 VALUES ($1, $2, $3, $4, $5, $6, 'aws')
			 ON CONFLICT (org_id, date, service, account_id, cloud_provider) DO UPDATE SET
			   amount = $5, currency = $6`,
			orgID, cost.Date, cost.Service, cost.AccountID, cost.Amount, cost.Currency,
		)
		if err != nil {
			log.Printf("Warning: failed to upsert AWS cost entry: %v", err)
		}
	}

	w.DB.Exec(ctx,
		"UPDATE aws_connections SET status = 'connected', last_sync_at = NOW(), error_message = '', updated_at = NOW() WHERE org_id = $1",
		orgID,
	)

	if err := w.runAIAnalysis(ctx, orgID); err != nil {
		log.Printf("Warning: AI analysis failed for org %s: %v", orgID, err)
	}

	return nil
}

func (w *CostSyncWorker) syncAzureCosts(ctx context.Context, orgID string) error {
	var tenantID, clientID, clientSecret, subscriptionID string
	err := w.DB.QueryRow(ctx,
		"SELECT tenant_id, client_id, client_secret, subscription_id FROM azure_connections WHERE org_id = $1",
		orgID,
	).Scan(&tenantID, &clientID, &clientSecret, &subscriptionID)
	if err != nil {
		return fmt.Errorf("no Azure connection found: %w", err)
	}

	// Get Azure bearer token
	authClient := azureclient.NewAuthClient(tenantID, clientID, clientSecret)
	tokenResp, err := authClient.GetToken("https://management.azure.com/.default")
	if err != nil {
		return fmt.Errorf("failed to authenticate with Azure: %w", err)
	}

	// Fetch costs
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	cmClient := azureclient.NewCostManagementClient()
	costs, err := cmClient.FetchDailyCosts(tokenResp.AccessToken, subscriptionID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch Azure costs: %w", err)
	}

	for _, cost := range costs {
		_, err := w.DB.Exec(ctx,
			`INSERT INTO daily_costs (org_id, date, service, account_id, amount, currency, cloud_provider)
			 VALUES ($1, $2, $3, $4, $5, $6, 'azure')
			 ON CONFLICT (org_id, date, service, account_id, cloud_provider) DO UPDATE SET
			   amount = $5, currency = $6`,
			orgID, cost.Date, cost.Service, cost.AccountID, cost.Amount, cost.Currency,
		)
		if err != nil {
			log.Printf("Warning: failed to upsert Azure cost entry: %v", err)
		}
	}

	w.DB.Exec(ctx,
		"UPDATE azure_connections SET status = 'connected', last_sync_at = NOW(), error_message = '', updated_at = NOW() WHERE org_id = $1",
		orgID,
	)

	if err := w.runAIAnalysis(ctx, orgID); err != nil {
		log.Printf("Warning: AI analysis failed for org %s: %v", orgID, err)
	}

	return nil
}

func (w *CostSyncWorker) runAIAnalysis(ctx context.Context, orgID string) error {
	log.Printf("🤖 Running AI analysis for org: %s", orgID)

	// --- Anomaly Detection ---
	rows, err := w.DB.Query(ctx,
		`SELECT date::text, service, amount FROM daily_costs 
		 WHERE org_id = $1 ORDER BY date ASC`,
		orgID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var dataPoints []ai.CostDataPoint
	for rows.Next() {
		var dp ai.CostDataPoint
		if rows.Scan(&dp.Date, &dp.Service, &dp.Amount) == nil {
			dataPoints = append(dataPoints, dp)
		}
	}

	anomalies := ai.DetectAnomalies(dataPoints)
	for _, a := range anomalies {
		narrative, _ := ai.GenerateAnomalyNarrative(ctx, w.Config, a)
		a.Narrative = narrative

		w.DB.Exec(ctx,
			`INSERT INTO anomalies (org_id, date, service, expected_amount, actual_amount, deviation_pct, confidence_score, narrative)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT DO NOTHING`,
			orgID, a.Date, a.Service, a.ExpectedAmount, a.ActualAmount, a.DeviationPct, a.ConfidenceScore, a.Narrative,
		)
	}

	// --- Forecasting ---
	trendRows, err := w.DB.Query(ctx,
		`SELECT date::text, SUM(amount) as total 
		 FROM daily_costs WHERE org_id = $1 
		 GROUP BY date ORDER BY date ASC`,
		orgID,
	)
	if err != nil {
		return err
	}
	defer trendRows.Close()

	var dailyTotals []ai.DailyTotal
	for trendRows.Next() {
		var dt ai.DailyTotal
		if trendRows.Scan(&dt.Date, &dt.Amount) == nil {
			dailyTotals = append(dailyTotals, dt)
		}
	}

	forecast := ai.ForecastMonthEnd(dailyTotals)
	if forecast != nil {
		var currentSpend float64
		for _, dt := range dailyTotals {
			currentSpend += dt.Amount
		}

		narrative, _ := ai.GenerateForecastNarrative(ctx, w.Config, forecast, currentSpend)
		forecast.Narrative = narrative

		w.DB.Exec(ctx,
			`INSERT INTO forecasts (org_id, forecast_date, predicted_total, best_case, worst_case, accuracy_pct, narrative)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			orgID, forecast.ForecastDate, forecast.PredictedTotal, forecast.BestCase, forecast.WorstCase,
			forecast.AccuracyPct, forecast.Narrative,
		)
	}

	log.Printf("✅ AI analysis complete for org: %s (anomalies: %d, forecast: %v)", orgID, len(anomalies), forecast != nil)
	return nil
}
