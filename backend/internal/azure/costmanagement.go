package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DailyCostEntry represents a single day+service cost row (matches AWS shape)
type DailyCostEntry struct {
	Date      string
	Service   string
	AccountID string // subscription ID for Azure
	Amount    float64
	Currency  string
}

// CostManagementClient calls the Azure Cost Management Query API
type CostManagementClient struct {
	httpClient *http.Client
}

// NewCostManagementClient creates a new cost management client
func NewCostManagementClient() *CostManagementClient {
	return &CostManagementClient{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// queryPayload is the request body for the Cost Management Query API
type queryPayload struct {
	Type       string         `json:"type"`
	Timeframe  string         `json:"timeframe"`
	TimePeriod *timePeriod    `json:"timePeriod,omitempty"`
	Dataset    datasetPayload `json:"dataset"`
}

type timePeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type datasetPayload struct {
	Granularity string             `json:"granularity"`
	Aggregation map[string]aggExpr `json:"aggregation"`
	Grouping    []groupExpr        `json:"grouping"`
}

type aggExpr struct {
	Name     string `json:"name"`
	Function string `json:"function"`
}

type groupExpr struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// queryResponse is the response from the Cost Management Query API
type queryResponse struct {
	Properties struct {
		NextLink string          `json:"nextLink"`
		Columns  []columnDef     `json:"columns"`
		Rows     [][]interface{} `json:"rows"`
	} `json:"properties"`
}

type columnDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// FetchDailyCosts fetches daily costs grouped by service for the last 30 days
func (c *CostManagementClient) FetchDailyCosts(accessToken, subscriptionID string, startDate, endDate time.Time) ([]DailyCostEntry, error) {
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.CostManagement/query?api-version=2023-11-01",
		subscriptionID,
	)

	payload := queryPayload{
		Type:      "ActualCost",
		Timeframe: "Custom",
		TimePeriod: &timePeriod{
			From: startDate.Format("2006-01-02T00:00:00Z"),
			To:   endDate.Format("2006-01-02T00:00:00Z"),
		},
		Dataset: datasetPayload{
			Granularity: "Daily",
			Aggregation: map[string]aggExpr{
				"totalCost": {Name: "Cost", Function: "Sum"},
			},
			Grouping: []groupExpr{
				{Type: "Dimension", Name: "ServiceName"},
			},
		},
	}

	var allEntries []DailyCostEntry
	currentURL := apiURL

	for {
		entries, nextLink, err := c.fetchPage(currentURL, accessToken, payload, subscriptionID)
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, entries...)

		if nextLink == "" {
			break
		}
		currentURL = nextLink
	}

	return allEntries, nil
}

func (c *CostManagementClient) fetchPage(apiURL, accessToken string, payload queryPayload, subscriptionID string) ([]DailyCostEntry, string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal query payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("cost management API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("cost management API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var queryResp queryResponse
	if err := json.Unmarshal(respBody, &queryResp); err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse column indices
	costIdx, serviceIdx, dateIdx := -1, -1, -1
	for i, col := range queryResp.Properties.Columns {
		switch col.Name {
		case "Cost":
			costIdx = i
		case "ServiceName":
			serviceIdx = i
		case "UsageDate":
			dateIdx = i
		}
	}

	var entries []DailyCostEntry
	for _, row := range queryResp.Properties.Rows {
		entry := DailyCostEntry{
			AccountID: subscriptionID,
			Currency:  "USD",
		}

		if costIdx >= 0 && costIdx < len(row) {
			if v, ok := row[costIdx].(float64); ok {
				entry.Amount = v
			}
		}

		if serviceIdx >= 0 && serviceIdx < len(row) {
			if v, ok := row[serviceIdx].(string); ok {
				entry.Service = v
			}
		}

		if dateIdx >= 0 && dateIdx < len(row) {
			// Azure returns date as number like 20260301 or string "20260301"
			switch v := row[dateIdx].(type) {
			case float64:
				dateStr := fmt.Sprintf("%.0f", v)
				if len(dateStr) == 8 {
					entry.Date = fmt.Sprintf("%s-%s-%s", dateStr[:4], dateStr[4:6], dateStr[6:8])
				}
			case string:
				if len(v) == 8 {
					entry.Date = fmt.Sprintf("%s-%s-%s", v[:4], v[4:6], v[6:8])
				} else {
					entry.Date = v
				}
			}
		}

		if entry.Service != "" && entry.Date != "" {
			entries = append(entries, entry)
		}
	}

	return entries, queryResp.Properties.NextLink, nil
}
