package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserController handles user-related API endpoints.
type UserController struct {
	userRepo repository.UserRepositoryInterface
}

func NewUserController(userRepo repository.UserRepositoryInterface) *UserController {
	return &UserController{userRepo: userRepo}
}

// Me handles GET /users/me — returns the currently authenticated user.
func (u *UserController) Me(c *gin.Context) {
	user, _ := c.Get(middleware.CurrentUserKey)
	c.JSON(http.StatusOK, user)
}

// List handles GET /users — returns all users with optional search/sort.
// Query params: q (username search), sort_by, sort_dir
func (u *UserController) List(c *gin.Context) {
	q := c.Query("q")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "ASC")

	users, err := u.userRepo.List(q, sortBy, sortDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": users, "total": len(users)})
}

// Get handles GET /users/:id — returns a single user by ID.
func (u *UserController) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	user, err := u.userRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Create handles POST /users — creates a new user.
// Body: {username, password, role_ids?, vendo_ids?, is_active?}
func (u *UserController) Create(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		RoleIDs  []uint `json:"role_ids"`
		VendoIDs []uint `json:"vendo_ids"`
		IsActive *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	isActive := true
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to hash password"})
		return
	}

	user := &models.User{
		Username: body.Username,
		Password: string(hash),
		IsActive: isActive,
	}
	if err := u.userRepo.Create(user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "username already exists or invalid data"})
		return
	}

	if len(body.RoleIDs) > 0 {
		if err := u.userRepo.AssignRoles(user.ID, body.RoleIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "user created but failed to assign roles"})
			return
		}
	}

	if len(body.VendoIDs) > 0 {
		if err := u.userRepo.AssignVendos(user.ID, body.VendoIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "user created but failed to assign vendos"})
			return
		}
	}

	created, _ := u.userRepo.GetByID(user.ID)
	c.JSON(http.StatusCreated, created)
}

// Update handles PUT /users/:id — updates an existing user.
// Body: {username?, password?, role_ids?, vendo_ids?, is_active?}
func (u *UserController) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	user, err := u.userRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "user not found"})
		return
	}

	var body struct {
		Username *string  `json:"username"`
		Password *string  `json:"password"`
		RoleIDs  *[]uint  `json:"role_ids"`
		VendoIDs *[]uint  `json:"vendo_ids"`
		IsActive *bool    `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if body.Username != nil {
		user.Username = *body.Username
	}
	if body.IsActive != nil {
		user.IsActive = *body.IsActive
	}

	// Clear associations before Save to avoid GORM preload conflicts.
	user.Roles = nil
	user.Vendos = nil

	if err := u.userRepo.Update(user); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "failed to update user"})
		return
	}

	if body.Password != nil && *body.Password != "" {
		if err := u.userRepo.UpdatePassword(user.ID, *body.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "user updated but failed to update password"})
			return
		}
	}

	if body.RoleIDs != nil {
		if err := u.userRepo.AssignRoles(user.ID, *body.RoleIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "user updated but failed to assign roles"})
			return
		}
	}

	if body.VendoIDs != nil {
		if err := u.userRepo.AssignVendos(user.ID, *body.VendoIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "user updated but failed to assign vendos"})
			return
		}
	}

	updated, _ := u.userRepo.GetByID(user.ID)
	c.JSON(http.StatusOK, updated)
}

// Delete handles DELETE /users/:id.
// Returns 422 if the caller tries to delete their own account.
func (u *UserController) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}

	currentUser := c.MustGet(middleware.CurrentUserKey).(*models.User)
	if id == currentUser.ID {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "cannot delete your own account"})
		return
	}

	if _, err := u.userRepo.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "user not found"})
		return
	}

	if err := u.userRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// ChangePassword handles PUT /users/me/password.
func (u *UserController) ChangePassword(c *gin.Context) {
	currentUser, _ := c.Get(middleware.CurrentUserKey)
	user := currentUser.(*models.User)

	var body struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "current_password and new_password are required"})
		return
	}

	if _, err := u.userRepo.CheckUser(user.Username, body.CurrentPassword); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "current password is incorrect"})
		return
	}

	if err := u.userRepo.UpdatePassword(user.ID, body.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

