package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TenantMiddleware(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("orgId")
		if orgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "organization ID required"})
			c.Abort()
			return
		}

		userID := c.GetString("user_id")

		// Verify user is a member of this org
		var role string
		err := db.QueryRow(context.Background(),
			"SELECT role FROM org_members WHERE org_id = $1 AND user_id = $2",
			orgID, userID,
		).Scan(&role)

		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
			c.Abort()
			return
		}

		c.Set("org_id", orgID)
		c.Set("org_role", role)
		c.Next()
	}
}
