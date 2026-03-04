package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	awsclient "github.com/finops/backend/internal/aws"
	"github.com/finops/backend/internal/config"
)

type AWSHandler struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Config *config.Config
}

type ConnectAWSRequest struct {
	RoleARN string `json:"role_arn" binding:"required"`
}

func (h *AWSHandler) Connect(c *gin.Context) {
	var req ConnectAWSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := c.GetString("org_id")
	role := c.GetString("org_role")

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only org owners can manage AWS connections"})
		return
	}

	// Generate unique external ID for this connection
	externalID := uuid.New().String()

	// Validate the role by attempting to assume it
	stsClient := awsclient.NewSTSClient(h.Config)
	err := stsClient.ValidateRole(c.Request.Context(), req.RoleARN, externalID)

	status := "connected"
	errMsg := ""
	if err != nil {
		status = "error"
		errMsg = err.Error()
	}

	// Upsert connection
	_, dbErr := h.DB.Exec(context.Background(),
		`INSERT INTO aws_connections (org_id, role_arn, external_id, status, error_message)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (org_id) DO UPDATE SET
		   role_arn = $2, external_id = $3, status = $4, error_message = $5, updated_at = NOW()`,
		orgID, req.RoleARN, externalID, status, errMsg,
	)

	if dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save connection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      status,
		"role_arn":    req.RoleARN,
		"external_id": externalID,
		"error":       errMsg,
	})
}

func (h *AWSHandler) Status(c *gin.Context) {
	orgID := c.GetString("org_id")

	var roleARN, externalID, status, errMsg string
	var lastSyncAt, createdAt, updatedAt interface{}

	err := h.DB.QueryRow(context.Background(),
		`SELECT role_arn, external_id, status, COALESCE(error_message, ''), last_sync_at, created_at, updated_at 
		 FROM aws_connections WHERE org_id = $1`,
		orgID,
	).Scan(&roleARN, &externalID, &status, &errMsg, &lastSyncAt, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"status":    "not_configured",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":    status == "connected",
		"status":       status,
		"role_arn":     roleARN,
		"external_id":  externalID,
		"error":        errMsg,
		"last_sync_at": lastSyncAt,
		"created_at":   createdAt,
	})
}

func (h *AWSHandler) TriggerSync(c *gin.Context) {
	orgID := c.GetString("org_id")
	role := c.GetString("org_role")

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only org owners can trigger sync"})
		return
	}

	// Verify connection exists and is healthy
	var status string
	err := h.DB.QueryRow(context.Background(),
		"SELECT status FROM aws_connections WHERE org_id = $1",
		orgID,
	).Scan(&status)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no AWS connection configured"})
		return
	}

	if status == "syncing" {
		c.JSON(http.StatusConflict, gin.H{"error": "sync already in progress"})
		return
	}

	// Update status to syncing
	h.DB.Exec(context.Background(),
		"UPDATE aws_connections SET status = 'syncing', updated_at = NOW() WHERE org_id = $1",
		orgID,
	)

	// Enqueue sync job in Redis
	err = h.Redis.RPush(context.Background(), "jobs:cost_sync", orgID).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue sync job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "cost sync job queued",
		"status":  "syncing",
	})
}
