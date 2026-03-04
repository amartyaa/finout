package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InsightsHandler struct {
	DB *pgxpool.Pool
}

func (h *InsightsHandler) Overview(c *gin.Context) {
	orgID := c.GetString("org_id")
	ctx := context.Background()

	// Current month boundaries
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	// Total spend this month
	var totalSpend float64
	h.DB.QueryRow(ctx,
		"SELECT COALESCE(SUM(amount), 0) FROM daily_costs WHERE org_id = $1 AND date >= $2 AND date < $3",
		orgID, monthStart, monthEnd,
	).Scan(&totalSpend)

	// Latest forecast
	var forecastTotal, bestCase, worstCase float64
	var forecastNarrative string
	err := h.DB.QueryRow(ctx,
		`SELECT predicted_total, best_case, worst_case, COALESCE(narrative, '') 
		 FROM forecasts WHERE org_id = $1 ORDER BY created_at DESC LIMIT 1`,
		orgID,
	).Scan(&forecastTotal, &bestCase, &worstCase, &forecastNarrative)
	if err != nil {
		forecastTotal = 0
		forecastNarrative = "No forecast available yet. Connect AWS and sync cost data to generate forecasts."
	}

	// Active anomalies count
	var anomalyCount int
	h.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM anomalies WHERE org_id = $1 AND status = 'active'",
		orgID,
	).Scan(&anomalyCount)

	// Potential savings
	var totalSavings float64
	h.DB.QueryRow(ctx,
		"SELECT COALESCE(SUM(estimated_monthly_savings), 0) FROM recommendations WHERE org_id = $1 AND status = 'active'",
		orgID,
	).Scan(&totalSavings)

	// Top 5 services this month
	rows, _ := h.DB.Query(ctx,
		`SELECT service, SUM(amount) as total 
		 FROM daily_costs 
		 WHERE org_id = $1 AND date >= $2 AND date < $3 
		 GROUP BY service 
		 ORDER BY total DESC LIMIT 5`,
		orgID, monthStart, monthEnd,
	)
	defer rows.Close()

	var topServices []gin.H
	for rows.Next() {
		var svc string
		var amt float64
		if rows.Scan(&svc, &amt) == nil {
			topServices = append(topServices, gin.H{"service": svc, "amount": amt})
		}
	}
	if topServices == nil {
		topServices = []gin.H{}
	}

	// Daily cost trend (last 30 days)
	trendRows, _ := h.DB.Query(ctx,
		`SELECT date, SUM(amount) as total 
		 FROM daily_costs 
		 WHERE org_id = $1 AND date >= $2
		 GROUP BY date 
		 ORDER BY date ASC`,
		orgID, now.AddDate(0, 0, -30),
	)
	defer trendRows.Close()

	var costTrend []gin.H
	for trendRows.Next() {
		var date time.Time
		var amt float64
		if trendRows.Scan(&date, &amt) == nil {
			costTrend = append(costTrend, gin.H{"date": date.Format("2006-01-02"), "amount": amt})
		}
	}
	if costTrend == nil {
		costTrend = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_spend_mtd": totalSpend,
		"forecast": gin.H{
			"predicted_total": forecastTotal,
			"best_case":       bestCase,
			"worst_case":      worstCase,
			"narrative":       forecastNarrative,
		},
		"anomaly_count":     anomalyCount,
		"potential_savings": totalSavings,
		"top_services":      topServices,
		"cost_trend":        costTrend,
	})
}

func (h *InsightsHandler) ListAnomalies(c *gin.Context) {
	orgID := c.GetString("org_id")

	rows, err := h.DB.Query(context.Background(),
		`SELECT id, date, service, expected_amount, actual_amount, deviation_pct, confidence_score, 
		        COALESCE(narrative, ''), status, created_at
		 FROM anomalies 
		 WHERE org_id = $1 
		 ORDER BY date DESC LIMIT 50`,
		orgID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch anomalies"})
		return
	}
	defer rows.Close()

	var anomalies []gin.H
	for rows.Next() {
		var id, service, narrative, status string
		var date time.Time
		var expected, actual, deviation, confidence float64
		var createdAt time.Time
		if rows.Scan(&id, &date, &service, &expected, &actual, &deviation, &confidence, &narrative, &status, &createdAt) == nil {
			anomalies = append(anomalies, gin.H{
				"id":               id,
				"date":             date.Format("2006-01-02"),
				"service":          service,
				"expected_amount":  expected,
				"actual_amount":    actual,
				"deviation_pct":    deviation,
				"confidence_score": confidence,
				"narrative":        narrative,
				"status":           status,
				"created_at":       createdAt,
			})
		}
	}
	if anomalies == nil {
		anomalies = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"anomalies": anomalies})
}

func (h *InsightsHandler) GetForecast(c *gin.Context) {
	orgID := c.GetString("org_id")

	var id, narrative string
	var forecastDate time.Time
	var predicted, best, worst, accuracy float64
	var createdAt time.Time

	err := h.DB.QueryRow(context.Background(),
		`SELECT id, forecast_date, predicted_total, best_case, worst_case, accuracy_pct, 
		        COALESCE(narrative, ''), created_at
		 FROM forecasts WHERE org_id = $1 ORDER BY created_at DESC LIMIT 1`,
		orgID,
	).Scan(&id, &forecastDate, &predicted, &best, &worst, &accuracy, &narrative, &createdAt)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   "No forecast available. Sync AWS cost data first.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available":       true,
		"id":              id,
		"forecast_date":   forecastDate.Format("2006-01-02"),
		"predicted_total": predicted,
		"best_case":       best,
		"worst_case":      worst,
		"accuracy_pct":    accuracy,
		"narrative":       narrative,
		"created_at":      createdAt,
	})
}

func (h *InsightsHandler) ListRecommendations(c *gin.Context) {
	orgID := c.GetString("org_id")

	rows, err := h.DB.Query(context.Background(),
		`SELECT id, category, resource_type, COALESCE(resource_id, ''), title, COALESCE(description, ''),
		        estimated_monthly_savings, risk_level, confidence_score, status, created_at
		 FROM recommendations 
		 WHERE org_id = $1 AND status = 'active'
		 ORDER BY estimated_monthly_savings DESC`,
		orgID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recommendations"})
		return
	}
	defer rows.Close()

	var recs []gin.H
	for rows.Next() {
		var id, category, resType, resID, title, desc, risk, status string
		var savings, confidence float64
		var createdAt time.Time
		if rows.Scan(&id, &category, &resType, &resID, &title, &desc, &savings, &risk, &confidence, &status, &createdAt) == nil {
			recs = append(recs, gin.H{
				"id":                        id,
				"category":                  category,
				"resource_type":             resType,
				"resource_id":               resID,
				"title":                     title,
				"description":               desc,
				"estimated_monthly_savings": savings,
				"risk_level":                risk,
				"confidence_score":          confidence,
				"status":                    status,
				"created_at":                createdAt,
			})
		}
	}
	if recs == nil {
		recs = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"recommendations": recs})
}

func (h *InsightsHandler) DismissRecommendation(c *gin.Context) {
	orgID := c.GetString("org_id")
	recID := c.Param("recId")

	tag, err := h.DB.Exec(context.Background(),
		"UPDATE recommendations SET status = 'dismissed', updated_at = NOW() WHERE id = $1 AND org_id = $2",
		recID, orgID,
	)

	if err != nil || tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "recommendation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recommendation dismissed"})
}
