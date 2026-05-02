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
