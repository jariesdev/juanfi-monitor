package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseUintParam reads a named path parameter as uint, writing a 400 response on error.
func parseUintParam(c *gin.Context, name string) (uint, error) {
	val, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid " + name})
	}
	return uint(val), err
}
