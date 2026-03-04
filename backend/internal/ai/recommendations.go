package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/finops/backend/internal/config"
)

type RecommendationResult struct {
	Category                string  `json:"category"`
	ResourceType            string  `json:"resource_type"`
	ResourceID              string  `json:"resource_id"`
	Title                   string  `json:"title"`
	Description             string  `json:"description"`
	EstimatedMonthlySavings float64 `json:"estimated_monthly_savings"`
	RiskLevel               string  `json:"risk_level"`
	ConfidenceScore         float64 `json:"confidence_score"`
}

type EC2Instance struct {
	InstanceID   string
	InstanceType string
	AvgCPU       float64
	MonthlyCost  float64
	Region       string
}

type EBSVolume struct {
	VolumeID    string
	SizeGB      int
	Attached    bool
	MonthlyCost float64
	Region      string
}

// AnalyzeEC2Utilization checks for underutilized EC2 instances
func AnalyzeEC2Utilization(instances []EC2Instance) []RecommendationResult {
	var recs []RecommendationResult

	for _, inst := range instances {
		if inst.AvgCPU < 20 {
			savings := inst.MonthlyCost * 0.4 // Estimate 40% savings from rightsizing
			risk := "low"
			if inst.AvgCPU < 5 {
				savings = inst.MonthlyCost * 0.7 // 70% for very idle instances
				risk = "low"                     // Low risk because it's barely used
			} else if inst.AvgCPU < 10 {
				savings = inst.MonthlyCost * 0.5
			}

			confidence := (20 - inst.AvgCPU) / 20 * 100

			recs = append(recs, RecommendationResult{
				Category:                "rightsizing",
				ResourceType:            "EC2",
				ResourceID:              inst.InstanceID,
				Title:                   fmt.Sprintf("Rightsize %s (%s)", inst.InstanceID, inst.InstanceType),
				EstimatedMonthlySavings: savings,
				RiskLevel:               risk,
				ConfidenceScore:         confidence,
			})
		}
	}

	return recs
}

// AnalyzeEBSVolumes checks for idle/unattached EBS volumes
func AnalyzeEBSVolumes(volumes []EBSVolume) []RecommendationResult {
	var recs []RecommendationResult

	for _, vol := range volumes {
		if !vol.Attached {
			recs = append(recs, RecommendationResult{
				Category:                "idle_resource",
				ResourceType:            "EBS",
				ResourceID:              vol.VolumeID,
				Title:                   fmt.Sprintf("Delete unattached EBS volume %s (%dGB)", vol.VolumeID, vol.SizeGB),
				EstimatedMonthlySavings: vol.MonthlyCost,
				RiskLevel:               "medium",
				ConfidenceScore:         95.0,
			})
		}
	}

	return recs
}

func GenerateRecommendationNarrative(ctx context.Context, cfg *config.Config, rec RecommendationResult) (string, error) {
	client := NewClient(cfg)

	systemPrompt := `You are a Cloud FinOps AI analyst. Generate a brief, actionable recommendation description.
Include the impact, risk, and specific next steps. Keep it under 100 words.`

	userPrompt := fmt.Sprintf(`Generate a recommendation description:
- Category: %s
- Resource: %s (%s)
- Resource ID: %s
- Estimated monthly savings: $%.2f
- Risk level: %s
- Confidence: %.0f%%

Explain what the issue is, why it matters, and what action to take.`,
		rec.Category, rec.ResourceType, rec.Title, rec.ResourceID,
		rec.EstimatedMonthlySavings, rec.RiskLevel, rec.ConfidenceScore)

	desc, err := client.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		desc = generateFallbackRecommendation(rec)
	}

	return strings.TrimSpace(desc), nil
}

func generateFallbackRecommendation(rec RecommendationResult) string {
	switch rec.Category {
	case "rightsizing":
		return fmt.Sprintf(
			"Instance %s is significantly underutilized. "+
				"Consider downsizing to a smaller instance type to save approximately $%.2f/month. "+
				"Risk is %s as the instance shows consistently low CPU usage.",
			rec.ResourceID, rec.EstimatedMonthlySavings, rec.RiskLevel,
		)
	case "idle_resource":
		return fmt.Sprintf(
			"EBS volume %s is not attached to any instance. "+
				"Deleting this unused volume would save $%.2f/month. "+
				"Verify no snapshots or data recovery is needed before deletion.",
			rec.ResourceID, rec.EstimatedMonthlySavings,
		)
	default:
		return fmt.Sprintf("Optimization opportunity with estimated savings of $%.2f/month.", rec.EstimatedMonthlySavings)
	}
}
