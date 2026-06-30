package controllers

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
)

var voucherPrefixPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]?$`)

// VendoVoucherController handles generated voucher listing and creation.
type VendoVoucherController struct {
	voucherRepo repository.VendoVoucherRepositoryInterface
	vendoRepo   repository.VendoRepositoryInterface
}

func NewVendoVoucherController(
	voucherRepo repository.VendoVoucherRepositoryInterface,
	vendoRepo repository.VendoRepositoryInterface,
) *VendoVoucherController {
	return &VendoVoucherController{
		voucherRepo: voucherRepo,
		vendoRepo:   vendoRepo,
	}
}

// ListForVendo handles GET /vendo-machines/:id/vouchers.
func (vc *VendoVoucherController) ListForVendo(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if !canAccessVendo(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "access denied"})
		return
	}

	vouchers, err := vc.voucherRepo.ListByVendo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vouchers})
}

// Generate handles POST /vendo-machines/:id/vouchers/generate.
func (vc *VendoVoucherController) Generate(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if !canAccessVendo(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "access denied"})
		return
	}

	var body struct {
		Prefix       string  `json:"prefix" binding:"required"`
		Amount       float64 `json:"amount" binding:"required"`
		Quantity     int     `json:"quantity" binding:"required"`
		AddToSales   bool    `json:"add_to_sales"`
		PrintThermal bool    `json:"print_thermal"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	if !voucherPrefixPattern.MatchString(body.Prefix) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "prefix must be 1-2 chars, starting with a letter"})
		return
	}
	if body.Quantity < 1 || body.Quantity > 15 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "quantity must be between 1 and 15"})
		return
	}
	if body.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "amount must be greater than 0"})
		return
	}

	vendo, err := vc.vendoRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}

	generated, err := services.NewJuanfiAPI(vendo).GenerateVouchers(
		body.Prefix,
		body.Amount,
		body.Quantity,
		body.AddToSales,
		body.PrintThermal,
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	vouchers := make([]models.VendoVoucher, 0, len(generated))
	for _, gv := range generated {
		vouchers = append(vouchers, models.VendoVoucher{
			VendoID:         id,
			Code:            gv.Code,
			Prefix:          body.Prefix,
			Amount:          gv.Amount,
			DurationMinutes: gv.Duration,
			AddedToSales:    body.AddToSales,
			PrintedThermal:  body.PrintThermal,
		})
	}

	if err := vc.voucherRepo.CreateBatch(vouchers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": vouchers})
}
