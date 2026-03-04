package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrgMember struct {
	ID       string    `json:"id"`
	OrgID    string    `json:"org_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Project struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Environment struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type AWSConnection struct {
	ID           string     `json:"id"`
	OrgID        string     `json:"org_id"`
	RoleARN      string     `json:"role_arn"`
	ExternalID   string     `json:"external_id"`
	Status       string     `json:"status"`
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type AzureConnection struct {
	ID             string     `json:"id"`
	OrgID          string     `json:"org_id"`
	TenantID       string     `json:"tenant_id"`
	ClientID       string     `json:"client_id"`
	ClientSecret   string     `json:"-"` // never expose in API responses
	SubscriptionID string     `json:"subscription_id"`
	Status         string     `json:"status"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type DailyCost struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	Date          string    `json:"date"`
	Service       string    `json:"service"`
	AccountID     string    `json:"account_id"`
	Environment   string    `json:"environment"`
	CloudProvider string    `json:"cloud_provider"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

type Anomaly struct {
	ID              string    `json:"id"`
	OrgID           string    `json:"org_id"`
	Date            string    `json:"date"`
	Service         string    `json:"service"`
	ExpectedAmount  float64   `json:"expected_amount"`
	ActualAmount    float64   `json:"actual_amount"`
	DeviationPct    float64   `json:"deviation_pct"`
	ConfidenceScore float64   `json:"confidence_score"`
	Narrative       string    `json:"narrative"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type Forecast struct {
	ID             string    `json:"id"`
	OrgID          string    `json:"org_id"`
	ForecastDate   string    `json:"forecast_date"`
	PredictedTotal float64   `json:"predicted_total"`
	BestCase       float64   `json:"best_case"`
	WorstCase      float64   `json:"worst_case"`
	AccuracyPct    float64   `json:"accuracy_pct"`
	Narrative      string    `json:"narrative"`
	CreatedAt      time.Time `json:"created_at"`
}

type Recommendation struct {
	ID                      string    `json:"id"`
	OrgID                   string    `json:"org_id"`
	Category                string    `json:"category"`
	ResourceType            string    `json:"resource_type"`
	ResourceID              string    `json:"resource_id"`
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	EstimatedMonthlySavings float64   `json:"estimated_monthly_savings"`
	RiskLevel               string    `json:"risk_level"`
	ConfidenceScore         float64   `json:"confidence_score"`
	Status                  string    `json:"status"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
