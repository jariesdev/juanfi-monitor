package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/authz"
	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
)

// VendoRateController handles per-vendo rate plans and the shared default
// template (rates with no vendo assigned).
type VendoRateController struct {
	rateRepo  repository.VendoRateRepositoryInterface
	vendoRepo repository.VendoRepositoryInterface
}

func NewVendoRateController(rateRepo repository.VendoRateRepositoryInterface, vendoRepo repository.VendoRepositoryInterface) *VendoRateController {
	return &VendoRateController{rateRepo: rateRepo, vendoRepo: vendoRepo}
}

// rateBody is the shared request shape for creating a rate (per-vendo or default).
type rateBody struct {
	Name         string  `json:"name" binding:"required"`
	Price        float64 `json:"price" binding:"required"`
	Minutes      int     `json:"minutes" binding:"required"`
	ValidityMins int     `json:"validity_minutes" binding:"required"`
	DataLimitMB  *int    `json:"data_limit_mb"`
	UserProfile  *string `json:"user_profile"`
}

// ListForVendo handles GET /vendo-machines/:id/rates.
func (rc *VendoRateController) ListForVendo(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if !authz.CanAccessVendo(c, id) {
		authz.AccessDenied(c, "access denied")
		return
	}
	rates, err := rc.rateRepo.ListByVendo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}

// Create handles POST /vendo-machines/:id/rates.
func (rc *VendoRateController) Create(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if !authz.CanAccessVendo(c, id) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "access denied"})
		return
	}

	var body rateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	rate := &models.VendoRate{
		VendoID:      &id,
		Name:         body.Name,
		Price:        body.Price,
		Minutes:      body.Minutes,
		ValidityMins: body.ValidityMins,
		DataLimitMB:  body.DataLimitMB,
		UserProfile:  body.UserProfile,
	}
	if err := rc.rateRepo.Create(rate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rate})
}

// ImportFromMachine handles POST /vendo-machines/:id/rates/import.
// It fetches the rate plans currently configured on the device and replaces
// the vendo's stored rates with them.
func (rc *VendoRateController) ImportFromMachine(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if !authz.CanAccessVendo(c, id) {
		authz.AccessDenied(c, "access denied")
		return
	}
	vendo, err := rc.vendoRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}

	deviceRates, err := services.NewJuanfiAPI(vendo).GetRates()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}

	rates := make([]models.VendoRate, len(deviceRates))
	for i, dr := range deviceRates {
		rates[i] = models.VendoRate{
			Name:         dr.Name,
			Price:        dr.Price,
			Minutes:      dr.Minutes,
			ValidityMins: dr.ValidityMins,
			DataLimitMB:  dr.DataLimitMB,
			UserProfile:  &dr.UserProfile,
			SortOrder:    i,
		}
	}
	if err := rc.rateRepo.ReplaceForVendo(id, rates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	saved, err := rc.rateRepo.ListByVendo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": saved})
}

// SetAsDefault handles POST /vendo-machines/:id/rates/set-as-default — admin only.
// It copies the vendo's current rates into the shared default template. The
// optional body field "mode" selects the merge strategy: "copy" appends to the
// existing default plan, anything else (the default) replaces it.
func (rc *VendoRateController) SetAsDefault(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if !authz.IsAdmin(c) {
		authz.AccessDenied(c, "admin access required")
		return
	}
	if _, err := rc.vendoRepo.GetByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}

	// Body is optional; an empty request defaults to replace semantics.
	var body struct {
		Mode string `json:"mode"`
	}
	_ = c.ShouldBindJSON(&body)

	rates, err := rc.rateRepo.ListByVendo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	if body.Mode == "copy" {
		err = rc.rateRepo.AppendToDefault(rates)
	} else {
		err = rc.rateRepo.ReplaceDefault(rates)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	saved, err := rc.rateRepo.ListDefault()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": saved})
}

// SyncToMachine handles POST /vendo-machines/:id/rates/sync.
// It uploads the vendo's stored rate plan to the physical device via the Juanfi
// saveRates API, overwriting whatever rates the machine currently has.
func (rc *VendoRateController) SyncToMachine(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}
	if !authz.CanAccessVendo(c, id) {
		authz.AccessDenied(c, "access denied")
		return
	}
	vendo, err := rc.vendoRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "vendo not found"})
		return
	}

	rates, err := rc.rateRepo.ListByVendo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	if len(rates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "no rates to sync"})
		return
	}

	if err := services.NewJuanfiAPI(vendo).SaveRates(toDeviceRates(rates)); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "synced", "count": len(rates)})
}

// toDeviceRates maps stored rate rows into the device-facing Rate shape.
// A missing/blank user profile becomes "default" — Mikrotik's default hotspot
// profile — matching how GetRates normalises rates read from the device.
func toDeviceRates(rates []models.VendoRate) []services.Rate {
	out := make([]services.Rate, len(rates))
	for i, r := range rates {
		profile := "default"
		if r.UserProfile != nil && *r.UserProfile != "" {
			profile = *r.UserProfile
		}
		out[i] = services.Rate{
			Name:         r.Name,
			Price:        r.Price,
			Minutes:      r.Minutes,
			ValidityMins: r.ValidityMins,
			DataLimitMB:  r.DataLimitMB,
			UserProfile:  profile,
		}
	}
	return out
}

// Update handles PUT /vendo-rates/:rateId.
func (rc *VendoRateController) Update(c *gin.Context) {
	id, err := parseUintParam(c, "rateId")
	if err != nil {
		return
	}
	rate, err := rc.rateRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "rate not found"})
		return
	}
	if !authz.CanManageRate(c, rate) {
		authz.AccessDenied(c, "access denied")
		return
	}

	var body struct {
		Name         *string  `json:"name"`
		Price        *float64 `json:"price"`
		Minutes      *int     `json:"minutes"`
		ValidityMins *int     `json:"validity_minutes"`
		DataLimitMB  *int     `json:"data_limit_mb"`
		UserProfile  *string  `json:"user_profile"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if body.Name != nil {
		rate.Name = *body.Name
	}
	if body.Price != nil {
		rate.Price = *body.Price
	}
	if body.Minutes != nil {
		rate.Minutes = *body.Minutes
	}
	if body.ValidityMins != nil {
		rate.ValidityMins = *body.ValidityMins
	}
	rate.DataLimitMB = body.DataLimitMB
	rate.UserProfile = body.UserProfile

	if err := rc.rateRepo.Update(rate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rate})
}

// Delete handles DELETE /vendo-rates/:rateId.
func (rc *VendoRateController) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "rateId")
	if err != nil {
		return
	}
	rate, err := rc.rateRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "rate not found"})
		return
	}
	if !authz.CanManageRate(c, rate) {
		authz.AccessDenied(c, "access denied")
		return
	}
	if err := rc.rateRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": nil})
}

// ListDefault handles GET /vendo-rates/default — admin only.
func (rc *VendoRateController) ListDefault(c *gin.Context) {
	if !authz.IsAdmin(c) {
		authz.AccessDenied(c, "admin access required")
		return
	}
	rates, err := rc.rateRepo.ListDefault()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}

// CreateDefault handles POST /vendo-rates/default — admin only.
func (rc *VendoRateController) CreateDefault(c *gin.Context) {
	if !authz.IsAdmin(c) {
		authz.AccessDenied(c, "admin access required")
		return
	}

	var body rateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	rate := &models.VendoRate{
		Name:         body.Name,
		Price:        body.Price,
		Minutes:      body.Minutes,
		ValidityMins: body.ValidityMins,
		DataLimitMB:  body.DataLimitMB,
		UserProfile:  body.UserProfile,
	}
	if err := rc.rateRepo.Create(rate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": rate})
}

// ApplyToAll handles POST /vendo-rates/apply-to-all — admin only.
// Clones the default template onto each given vendo, replacing that vendo's
// existing rates. The frontend always sends the full target list (defaulted
// to every vendo) so there's no implicit "empty means all" behavior here.
func (rc *VendoRateController) ApplyToAll(c *gin.Context) {
	if !authz.IsAdmin(c) {
		authz.AccessDenied(c, "admin access required")
		return
	}

	var body struct {
		VendoIDs []uint `json:"vendo_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	if err := rc.rateRepo.ApplyDefaultToVendos(body.VendoIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "applied", "count": len(body.VendoIDs)})
}
