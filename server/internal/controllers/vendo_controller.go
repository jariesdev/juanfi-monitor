package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
)

// VendoController handles all vendo machine CRUD and operation endpoints.
type VendoController struct {
	vendoRepo      repository.VendoRepositoryInterface
	withdrawalRepo repository.WithdrawalRepositoryInterface
	userRepo       repository.UserRepositoryInterface
}

func NewVendoController(vendoRepo repository.VendoRepositoryInterface, withdrawalRepo repository.WithdrawalRepositoryInterface, userRepo repository.UserRepositoryInterface) *VendoController {
	return &VendoController{
		vendoRepo:      vendoRepo,
		withdrawalRepo: withdrawalRepo,
		userRepo:       userRepo,
	}
}

// All handles GET /vendo-machines — list/search vendo machines.
// Non-admin users (lacking PermUsers) only see their assigned vendos.
// Query params: q (name search), is_active (bool filter).
func (v *VendoController) All(c *gin.Context) {
	q := c.Query("q")
	isActiveStr := c.Query("is_active")

	var qPtr *string
	if q != "" {
		qPtr = &q
	}

	var isActivePtr *bool
	if isActiveStr != "" {
		b, err := strconv.ParseBool(isActiveStr)
		if err == nil {
			isActivePtr = &b
		}
	}

	currentUser := c.MustGet(middleware.CurrentUserKey).(*models.User)
	var assignedIDs []uint
	if !currentUser.HasPermission(models.PermUsers) {
		ids, err := v.userRepo.GetVendoIDs(currentUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		assignedIDs = ids
	}

	vendos, err := v.vendoRepo.Search(qPtr, isActivePtr, assignedIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vendos})
}

// Get handles GET /vendo-machines/:id — retrieve a single vendo.
func (v *VendoController) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	vendo, err := v.vendoRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vendo})
}

// Store handles POST /vendo-machines — create a new vendo machine.
func (v *VendoController) Store(c *gin.Context) {
	var body struct {
		Name       string  `json:"name" binding:"required"`
		MacAddress *string `json:"mac_address"`
		APIURL     *string `json:"api_url"`
		APIKey     *string `json:"api_key"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	vendo := &models.Vendo{
		Name:       body.Name,
		MacAddress: body.MacAddress,
		APIURL:     body.APIURL,
		APIKey:     body.APIKey,
	}
	if err := v.vendoRepo.Create(vendo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vendo})
}

// Delete handles DELETE /vendo-machines/:id.
func (v *VendoController) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if err := v.vendoRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": nil})
}

// Status handles GET /vendo-machines/:id/status — live status from the machine.
func (v *VendoController) Status(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	vendo, err := v.vendoRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}

	status, err := services.NewJuanfiAPI(vendo).GetSystemStatus()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// Withdraw handles POST /vendo-machines/:id/withdraw-current-sales.
// It reads the current coin count, resets it on the machine, and records the withdrawal.
func (v *VendoController) Withdraw(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	vendo, err := v.vendoRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}

	api := services.NewJuanfiAPI(vendo)
	status, err := api.GetSystemStatus()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	if err := api.ResetCurrentSales(); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	amount := float64(status.CurrentCoinCount)
	withdrawal, err := v.withdrawalRepo.Add(vendo.ID, amount, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": withdrawal})
}

// ActiveUsers handles GET /vendo-machines/:id/active-users — live connected users.
func (v *VendoController) ActiveUsers(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	vendo, err := v.vendoRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}

	users, err := services.NewJuanfiAPI(vendo).GetActiveUsers()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// SetStatus handles POST /vendo-machines/:id/set-status.
// Request body: {"status": true|false}
func (v *VendoController) SetStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	var body struct {
		Status bool `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if err := v.vendoRepo.SetStatus(id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// parseID is a helper that reads and validates the :id path parameter.
func parseID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid id"})
		return 0, err
	}
	return uint(id), nil
}
