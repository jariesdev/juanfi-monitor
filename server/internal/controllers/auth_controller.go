package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// jwtClaims embeds the standard RegisteredClaims (exp, iat, etc.) and adds
// the username so we can look up the user on every authenticated request.
type jwtClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthController handles login and JWT token issuance.
type AuthController struct {
	userRepo  repository.UserRepositoryInterface
	jwtSecret string
}

func NewAuthController(userRepo repository.UserRepositoryInterface, jwtSecret string) *AuthController {
	return &AuthController{userRepo: userRepo, jwtSecret: jwtSecret}
}

// Login handles POST /token.
// Accepts application/x-www-form-urlencoded (OAuth2 password flow).
// Returns a signed HS256 JWT. Response shape is identical to the previous
// implementation so the SvelteKit frontend requires no changes.
func (a *AuthController) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := a.userRepo.CheckUser(username, password)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Incorrect username or password"})
		return
	}

	expiresAt := time.Now().Add(1 * time.Hour)

	claims := jwtClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(a.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"expiry":       expiresAt.Unix(), // Unix timestamp (number), same as Python
		"token_type":   "bearer",
		"user":         user,
	})
}
