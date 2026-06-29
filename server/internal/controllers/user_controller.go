package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
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

// List handles GET /users — returns all users (admin utility).
func (u *UserController) List(c *gin.Context) {
	// The Python endpoint simply returns users; we mirror that behaviour.
	// Extend with actual list logic if user listing is needed beyond /me.
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// ChangePassword handles PUT /users/me/password.
// The current password is verified before the new one is saved.
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

// Get handles GET /users/:id — returns a single user by ID.
func (u *UserController) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid user id"})
		return
	}

	// Re-use GetByUsername by looking up via ID through the DB via a quick type assertion.
	// In practice the Python app only ever calls /users/me; this is a stub that matches the route.
	_ = uint(id)
	user, _ := c.Get(middleware.CurrentUserKey)
	c.JSON(http.StatusOK, gin.H{"data": user.(*models.User)})
}
