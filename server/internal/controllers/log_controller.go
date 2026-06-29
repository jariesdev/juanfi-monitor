package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
	"gorm.io/gorm"
)

// LogController handles vendo system log endpoints.
type LogController struct {
	logRepo   repository.LogRepositoryInterface
	vendoRepo repository.VendoRepositoryInterface
	db        *gorm.DB
}

func NewLogController(db *gorm.DB, logRepo repository.LogRepositoryInterface, vendoRepo repository.VendoRepositoryInterface) *LogController {
	return &LogController{db: db, logRepo: logRepo, vendoRepo: vendoRepo}
}

// Search handles GET /logs — paginated log search.
// Query params: q (description filter), date (YYYY-MM-DD), vendo_id, page, size.
func (l *LogController) Search(c *gin.Context) {
	q := c.Query("q")
	date := c.Query("date")
	vendoIDStr := c.Query("vendo_id")
	page, size := paginationParams(c)

	var qPtr *string
	if q != "" {
		qPtr = &q
	}
	var datePtr *string
	if date != "" {
		datePtr = &date
	}
	var vendoIDPtr *uint
	if vendoIDStr != "" {
		if id, err := strconv.ParseUint(vendoIDStr, 10, 64); err == nil {
			uid := uint(id)
			vendoIDPtr = &uid
		}
	}

	result, err := l.logRepo.Search(qPtr, datePtr, vendoIDPtr, assignedVendoIDs(c), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Refresh handles POST /log/refresh — manually triggers a log pull for all active vendos.
func (l *LogController) Refresh(c *gin.Context) {
	vendos, err := l.vendoRepo.AllActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	for i := range vendos {
		v := &vendos[i]
		logger := services.NewJuanfiLogger(v, l.db)
		if err := logger.Run(); err != nil {
			log.Printf("refresh logs [%s]: %v", v.Name, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// paginationParams extracts page and size from query params with sane defaults.
func paginationParams(c *gin.Context) (page, size int) {
	page = 1
	size = 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if s := c.Query("size"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= 100 {
			size = v
		}
	}
	return
}
