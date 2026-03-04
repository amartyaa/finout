package ai

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/finops/backend/internal/config"
)

type AnomalyResult struct {
	Date            string  `json:"date"`
	Service         string  `json:"service"`
	ExpectedAmount  float64 `json:"expected_amount"`
	ActualAmount    float64 `json:"actual_amount"`
	DeviationPct    float64 `json:"deviation_pct"`
	ConfidenceScore float64 `json:"confidence_score"`
	Narrative       string  `json:"narrative"`
}

type CostDataPoint struct {
	Date    string
	Service string
	Amount  float64
}

func DetectAnomalies(data []CostDataPoint) []AnomalyResult {
	// Group by service
	serviceData := make(map[string][]CostDataPoint)
	for _, d := range data {
		serviceData[d.Service] = append(serviceData[d.Service], d)
	}

	var anomalies []AnomalyResult

	for service, points := range serviceData {
		if len(points) < 7 {
			continue // Need at least 7 days of data
		}

		// Calculate rolling 14-day mean and stddev (or all available data if < 14)
		windowSize := 14
		if len(points) < windowSize {
			windowSize = len(points) - 1
		}

		for i := windowSize; i < len(points); i++ {
			window := points[i-windowSize : i]

			mean := 0.0
			for _, p := range window {
				mean += p.Amount
			}
			mean /= float64(len(window))

			variance := 0.0
			for _, p := range window {
				diff := p.Amount - mean
				variance += diff * diff
			}
			variance /= float64(len(window))
			stddev := math.Sqrt(variance)

			actual := points[i].Amount
			if stddev == 0 {
				continue
			}

			deviation := (actual - mean) / stddev

			if deviation > 2.0 { // More than 2 sigma
				deviationPct := ((actual - mean) / mean) * 100
				confidence := math.Min(deviation/4.0, 1.0) * 100 // Scale to 0-100

				anomalies = append(anomalies, AnomalyResult{
					Date:            points[i].Date,
					Service:         service,
					ExpectedAmount:  math.Round(mean*100) / 100,
					ActualAmount:    math.Round(actual*100) / 100,
					DeviationPct:    math.Round(deviationPct*100) / 100,
					ConfidenceScore: math.Round(confidence*100) / 100,
				})
			}
		}
	}

	return anomalies
}

func GenerateAnomalyNarrative(ctx context.Context, cfg *config.Config, anomaly AnomalyResult) (string, error) {
	client := NewClient(cfg)

	systemPrompt := `You are a Cloud FinOps AI analyst. Generate a brief, clear narrative explaining a cost anomaly.
Be specific about what likely caused the spike. Keep it under 150 words.
Use plain English suitable for a business dashboard.`

	userPrompt := fmt.Sprintf(`Cost anomaly detected:
- Date: %s
- Service: %s
- Expected daily cost: $%.2f
- Actual cost: $%.2f  
- Deviation: %.1f%% above normal
- Confidence: %.0f%%

Explain what likely caused this spike and what the user should investigate.`,
		anomaly.Date, anomaly.Service, anomaly.ExpectedAmount, anomaly.ActualAmount,
		anomaly.DeviationPct, anomaly.ConfidenceScore)

	narrative, err := client.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		// Fallback to template-based narrative if LLM fails
		narrative = generateFallbackAnomalyNarrative(anomaly)
	}

	return strings.TrimSpace(narrative), nil
}

func generateFallbackAnomalyNarrative(a AnomalyResult) string {
	return fmt.Sprintf(
		"A cost spike of %.1f%% was detected for %s on %s. "+
			"Daily spend jumped from an expected $%.2f to $%.2f. "+
			"This is a significant deviation that warrants investigation. "+
			"Check for increased resource usage, new deployments, or configuration changes.",
		a.DeviationPct, a.Service, a.Date, a.ExpectedAmount, a.ActualAmount,
	)
}
