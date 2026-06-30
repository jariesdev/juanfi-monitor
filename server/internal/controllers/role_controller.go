package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// RoleController handles CRUD operations for roles.
type RoleController struct {
	roleRepo repository.RoleRepositoryInterface
}

func NewRoleController(roleRepo repository.RoleRepositoryInterface) *RoleController {
	return &RoleController{roleRepo: roleRepo}
}

// List handles GET /roles — returns all roles.
func (r *RoleController) List(c *gin.Context) {
	roles, err := r.roleRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// Get handles GET /roles/:id.
func (r *RoleController) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	role, err := r.roleRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// Create handles POST /roles.
// Body: {name, permissions: []string}
func (r *RoleController) Create(c *gin.Context) {
	var body struct {
		Name        string   `json:"name" binding:"required"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	role := &models.Role{Name: body.Name}
	role.SetPermissions(body.Permissions)

	if err := r.roleRepo.Create(role); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "role name already exists or invalid data"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// Update handles PUT /roles/:id.
func (r *RoleController) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	role, err := r.roleRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "role not found"})
		return
	}

	var body struct {
		Name        *string  `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if body.Name != nil {
		role.Name = *body.Name
	}
	if body.Permissions != nil {
		role.SetPermissions(body.Permissions)
	}

	if err := r.roleRepo.Update(role); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// Delete handles DELETE /roles/:id.
// Returns 422 if the role is still assigned to any users.
func (r *RoleController) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	if _, err := r.roleRepo.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "role not found"})
		return
	}

	count, err := r.roleRepo.CountUsersWithRole(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if count > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "cannot delete a role that is assigned to users"})
		return
	}

	if err := r.roleRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}
