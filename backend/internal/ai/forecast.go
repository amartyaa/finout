package ai

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/finops/backend/internal/config"
)

type ForecastResult struct {
	ForecastDate   string  `json:"forecast_date"`
	PredictedTotal float64 `json:"predicted_total"`
	BestCase       float64 `json:"best_case"`
	WorstCase      float64 `json:"worst_case"`
	AccuracyPct    float64 `json:"accuracy_pct"`
	Narrative      string  `json:"narrative"`
}

type DailyTotal struct {
	Date   string
	Amount float64
}

// ForecastMonthEnd uses linear regression on daily totals to predict end-of-month spend
func ForecastMonthEnd(dailyTotals []DailyTotal) *ForecastResult {
	if len(dailyTotals) < 7 {
		return nil
	}

	// Linear regression: y = mx + b
	n := float64(len(dailyTotals))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, d := range dailyTotals {
		x := float64(i)
		y := d.Amount
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return nil
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / n

	// Calculate R² for accuracy
	meanY := sumY / n
	ssTot, ssRes := 0.0, 0.0
	for i, d := range dailyTotals {
		predicted := slope*float64(i) + intercept
		ssTot += (d.Amount - meanY) * (d.Amount - meanY)
		ssRes += (d.Amount - predicted) * (d.Amount - predicted)
	}

	rSquared := 0.0
	if ssTot > 0 {
		rSquared = 1 - (ssRes / ssTot)
	}

	// Calculate standard error for confidence bands
	stdErr := 0.0
	if n > 2 {
		stdErr = math.Sqrt(ssRes / (n - 2))
	}

	// Predict month-end total
	now := time.Now()
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	currentDay := now.Day()
	remainingDays := daysInMonth - currentDay

	// Sum current actual spend
	currentTotal := 0.0
	for _, d := range dailyTotals {
		currentTotal += d.Amount
	}

	// Project remaining days
	projectedRemaining := 0.0
	for i := 0; i < remainingDays; i++ {
		day := float64(len(dailyTotals) + i)
		projectedRemaining += slope*day + intercept
	}

	predicted := currentTotal + projectedRemaining
	bestCase := currentTotal + projectedRemaining*(1-stdErr/meanY)
	worstCase := currentTotal + projectedRemaining*(1+stdErr/meanY)

	// Ensure no negative values
	predicted = math.Max(predicted, 0)
	bestCase = math.Max(bestCase, 0)
	worstCase = math.Max(worstCase, 0)

	return &ForecastResult{
		ForecastDate:   now.Format("2006-01-02"),
		PredictedTotal: math.Round(predicted*100) / 100,
		BestCase:       math.Round(bestCase*100) / 100,
		WorstCase:      math.Round(worstCase*100) / 100,
		AccuracyPct:    math.Round(rSquared*10000) / 100, // As percentage
	}
}

func GenerateForecastNarrative(ctx context.Context, cfg *config.Config, forecast *ForecastResult, currentSpend float64) (string, error) {
	client := NewClient(cfg)

	systemPrompt := `You are a Cloud FinOps AI analyst. Generate a brief, clear forecast summary.
Highlight trends, risks, and actionable insights. Keep it under 150 words. Use plain English.`

	userPrompt := fmt.Sprintf(`Monthly cloud spend forecast:
- Current month-to-date spend: $%.2f
- Predicted end-of-month total: $%.2f
- Best case: $%.2f
- Worst case: $%.2f
- Model accuracy (R²): %.1f%%
- Forecast date: %s

Provide a clear forecast summary with any risk warnings.`,
		currentSpend, forecast.PredictedTotal, forecast.BestCase, forecast.WorstCase,
		forecast.AccuracyPct, forecast.ForecastDate)

	narrative, err := client.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		narrative = fmt.Sprintf(
			"Based on current spending trends, your projected end-of-month cloud cost is $%.2f. "+
				"This falls within a range of $%.2f (best case) to $%.2f (worst case). "+
				"The forecast model shows %.1f%% accuracy.",
			forecast.PredictedTotal, forecast.BestCase, forecast.WorstCase, forecast.AccuracyPct,
		)
	}

	return strings.TrimSpace(narrative), nil
}
