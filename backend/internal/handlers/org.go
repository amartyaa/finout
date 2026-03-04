package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrgHandler struct {
	DB *pgxpool.Pool
}

type CreateOrgRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateEnvRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *OrgHandler) CreateOrg(c *gin.Context) {
	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	slug := generateSlug(req.Name)

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}
	defer tx.Rollback(ctx)

	var orgID string
	err = tx.QueryRow(ctx,
		"INSERT INTO organizations (name, slug, created_by) VALUES ($1, $2, $3) RETURNING id",
		req.Name, slug, userID,
	).Scan(&orgID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "organization name already taken"})
		return
	}

	// Add creator as owner
	_, err = tx.Exec(ctx,
		"INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'owner')",
		orgID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add owner"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   orgID,
		"name": req.Name,
		"slug": slug,
		"role": "owner",
	})
}

func (h *OrgHandler) ListOrgs(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.DB.Query(context.Background(),
		`SELECT o.id, o.name, o.slug, om.role, o.created_at 
		 FROM organizations o 
		 JOIN org_members om ON o.id = om.org_id 
		 WHERE om.user_id = $1 
		 ORDER BY o.created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list organizations"})
		return
	}
	defer rows.Close()

	var orgs []gin.H
	for rows.Next() {
		var id, name, slug, role string
		var createdAt interface{}
		if err := rows.Scan(&id, &name, &slug, &role, &createdAt); err != nil {
			continue
		}
		orgs = append(orgs, gin.H{
			"id":         id,
			"name":       name,
			"slug":       slug,
			"role":       role,
			"created_at": createdAt,
		})
	}

	if orgs == nil {
		orgs = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}

func (h *OrgHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orgID := c.GetString("org_id")

	var projectID string
	err := h.DB.QueryRow(context.Background(),
		"INSERT INTO projects (org_id, name, description) VALUES ($1, $2, $3) RETURNING id",
		orgID, req.Name, req.Description,
	).Scan(&projectID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "project name already exists in this org"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          projectID,
		"org_id":      orgID,
		"name":        req.Name,
		"description": req.Description,
	})
}

func (h *OrgHandler) ListProjects(c *gin.Context) {
	orgID := c.GetString("org_id")

	rows, err := h.DB.Query(context.Background(),
		`SELECT id, name, description, created_at FROM projects WHERE org_id = $1 ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list projects"})
		return
	}
	defer rows.Close()

	var projects []gin.H
	for rows.Next() {
		var id, name, desc string
		var createdAt interface{}
		if err := rows.Scan(&id, &name, &desc, &createdAt); err != nil {
			continue
		}
		projects = append(projects, gin.H{
			"id":          id,
			"name":        name,
			"description": desc,
			"created_at":  createdAt,
		})
	}

	if projects == nil {
		projects = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func (h *OrgHandler) CreateEnvironment(c *gin.Context) {
	var req CreateEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validEnvs := map[string]bool{"dev": true, "staging": true, "prod": true, "test": true, "custom": true}
	if !validEnvs[req.Name] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid environment name, must be one of: dev, staging, prod, test, custom"})
		return
	}

	projectID := c.Param("projectId")
	orgID := c.GetString("org_id")

	// Verify project belongs to org
	var exists bool
	h.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND org_id = $2)",
		projectID, orgID,
	).Scan(&exists)

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var envID string
	err := h.DB.QueryRow(context.Background(),
		"INSERT INTO environments (project_id, name) VALUES ($1, $2) RETURNING id",
		projectID, req.Name,
	).Scan(&envID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "environment already exists for this project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         envID,
		"project_id": projectID,
		"name":       req.Name,
	})
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
