// Package middleware provides Gin middleware functions.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jariesdev/vendoreport/internal/repository"
)

const CurrentUserKey = "current_user"

// jwtClaims mirrors the claims structure issued by AuthController.
type jwtClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Auth validates the HS256 JWT Bearer token from the Authorization header.
// On success it injects the *models.User into the Gin context under CurrentUserKey.
// The library handles expiry validation automatically via RegisteredClaims.ExpiresAt.
func Auth(userRepo repository.UserRepositoryInterface, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "missing or invalid authorization header"})
			return
		}

		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(rawToken, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			// Reject tokens signed with any algorithm other than HS256.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(*jwtClaims)
		if !ok || claims.Username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "invalid token claims"})
			return
		}

		user, err := userRepo.GetByUsername(claims.Username)
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "user not found"})
			return
		}

		c.Set(CurrentUserKey, user)
		c.Next()
	}
}
