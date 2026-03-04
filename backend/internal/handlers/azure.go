package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	azureclient "github.com/finops/backend/internal/azure"
)

type AzureHandler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

type ConnectAzureRequest struct {
	TenantID       string `json:"tenant_id" binding:"required"`
	ClientID       string `json:"client_id" binding:"required"`
	ClientSecret   string `json:"client_secret" binding:"required"`
	SubscriptionID string `json:"subscription_id" binding:"required"`
}

func (h *AzureHandler) Connect(c *gin.Context) {
	var req ConnectAzureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := c.GetString("org_id")
	role := c.GetString("org_role")

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only org owners can manage Azure connections"})
		return
	}

	// Validate credentials by attempting to get a token
	authClient := azureclient.NewAuthClient(req.TenantID, req.ClientID, req.ClientSecret)
	err := authClient.ValidateCredentials()

	status := "connected"
	errMsg := ""
	if err != nil {
		status = "error"
		errMsg = err.Error()
	}

	// Upsert connection
	_, dbErr := h.DB.Exec(context.Background(),
		`INSERT INTO azure_connections (org_id, tenant_id, client_id, client_secret, subscription_id, status, error_message)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (org_id) DO UPDATE SET
		   tenant_id = $2, client_id = $3, client_secret = $4, subscription_id = $5, 
		   status = $6, error_message = $7, updated_at = NOW()`,
		orgID, req.TenantID, req.ClientID, req.ClientSecret, req.SubscriptionID, status, errMsg,
	)

	if dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          status,
		"tenant_id":       req.TenantID,
		"subscription_id": req.SubscriptionID,
		"error":           errMsg,
	})
}

func (h *AzureHandler) Status(c *gin.Context) {
	orgID := c.GetString("org_id")

	var tenantID, clientID, subscriptionID, status, errMsg string
	var lastSyncAt, createdAt, updatedAt interface{}

	err := h.DB.QueryRow(context.Background(),
		`SELECT tenant_id, client_id, subscription_id, status, COALESCE(error_message, ''), last_sync_at, created_at, updated_at 
		 FROM azure_connections WHERE org_id = $1`,
		orgID,
	).Scan(&tenantID, &clientID, &subscriptionID, &status, &errMsg, &lastSyncAt, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"status":    "not_configured",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":       status == "connected",
		"status":          status,
		"tenant_id":       tenantID,
		"client_id":       clientID,
		"subscription_id": subscriptionID,
		"error":           errMsg,
		"last_sync_at":    lastSyncAt,
		"created_at":      createdAt,
	})
}

func (h *AzureHandler) TriggerSync(c *gin.Context) {
	orgID := c.GetString("org_id")
	role := c.GetString("org_role")

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only org owners can trigger sync"})
		return
	}

	// Verify connection exists and is healthy
	var status string
	err := h.DB.QueryRow(context.Background(),
		"SELECT status FROM azure_connections WHERE org_id = $1",
		orgID,
	).Scan(&status)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no Azure connection configured"})
		return
	}

	if status == "syncing" {
		c.JSON(http.StatusConflict, gin.H{"error": "sync already in progress"})
		return
	}

	// Update status to syncing
	h.DB.Exec(context.Background(),
		"UPDATE azure_connections SET status = 'syncing', updated_at = NOW() WHERE org_id = $1",
		orgID,
	)

	// Enqueue sync job in Redis
	err = h.Redis.RPush(context.Background(), "jobs:azure_cost_sync", orgID).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue sync job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Azure cost sync job queued",
		"status":  "syncing",
	})
}
