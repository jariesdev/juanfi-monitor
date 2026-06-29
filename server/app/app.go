// Package app wires together all dependencies and returns a configured Gin engine.
// Both cmd/main.go and integration tests import this package so the router
// is constructed from a single source of truth.
package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/internal/controllers"
	"github.com/jariesdev/vendoreport/internal/middleware"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/scheduler"
	ws "github.com/jariesdev/vendoreport/internal/websocket"
	"gorm.io/gorm"
)

// App holds the live components that need lifecycle management (cron, ws hub).
type App struct {
	Router    *gin.Engine
	Hub       *ws.Hub
	// StopScheduler stops the background cron jobs. Call it on shutdown.
	StopScheduler func()
}

// New wires all repositories, controllers, and middleware, then returns an App.
// jwtSecret is the HS256 signing key used to issue and validate Bearer tokens.
// isProd disables request logging to avoid the per-request stdout overhead in production.
// trustedProxies is the list of IPs/CIDRs of upstream proxies (e.g. HAProxy) so Gin
// reads X-Forwarded-For correctly; pass nil to trust no proxies.
// Pass startScheduler=false in tests to skip cron jobs.
func New(db *gorm.DB, corsOrigins []string, jwtSecret string, isProd bool, startScheduler bool, trustedProxies []string) *App {
	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	vendoRepo := repository.NewVendoRepository(db)
	logRepo := repository.NewLogRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	statusRepo := repository.NewVendoStatusRepository(db)
	withdrawalRepo := repository.NewWithdrawalRepository(db)

	// ── Controllers ──────────────────────────────────────────────────────────
	authCtrl := controllers.NewAuthController(userRepo, jwtSecret)
	userCtrl := controllers.NewUserController(userRepo)
	vendoCtrl := controllers.NewVendoController(vendoRepo, withdrawalRepo)
	logCtrl := controllers.NewLogController(db, logRepo, vendoRepo)
	saleCtrl := controllers.NewSaleController(saleRepo)
	statusCtrl := controllers.NewVendoStatusController(statusRepo)
	withdrawalCtrl := controllers.NewWithdrawalController(withdrawalRepo)

	// ── WebSocket Hub ─────────────────────────────────────────────────────────
	hub := ws.NewHub()
	go hub.Run()

	// ── Cron Scheduler ────────────────────────────────────────────────────────
	var stopFn func()
	if startScheduler {
		c := scheduler.Start(db, hub)
		stopFn = func() { c.Stop() }
	} else {
		stopFn = func() {}
	}

	// ── Router ────────────────────────────────────────────────────────────────
	router := buildRouter(corsOrigins, jwtSecret, isProd, trustedProxies, hub, authCtrl, userCtrl, vendoCtrl, logCtrl, saleCtrl, statusCtrl, withdrawalCtrl, userRepo)

	return &App{Router: router, Hub: hub, StopScheduler: stopFn}
}

func buildRouter(
	allowedOrigins []string,
	jwtSecret string,
	isProd bool,
	trustedProxies []string,
	hub *ws.Hub,
	authCtrl *controllers.AuthController,
	userCtrl *controllers.UserController,
	vendoCtrl *controllers.VendoController,
	logCtrl *controllers.LogController,
	saleCtrl *controllers.SaleController,
	statusCtrl *controllers.VendoStatusController,
	withdrawalCtrl *controllers.WithdrawalController,
	userRepo repository.UserRepositoryInterface,
) *gin.Engine {
	router := gin.New()
	router.SetTrustedProxies(trustedProxies)
	if !isProd {
		// Request logging is skipped in production to avoid per-request stdout overhead.
		router.Use(gin.Logger())
	}
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(allowedOrigins))

	// Public routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "VendoReport API"})
	})
	router.POST("/token", authCtrl.Login)
	router.GET("/ws", ws.Handler(hub))

	// Protected routes
	auth := router.Group("/")
	auth.Use(middleware.Auth(userRepo, jwtSecret))

	auth.GET("/users/me", userCtrl.Me)
	auth.GET("/users", userCtrl.List)
	auth.GET("/users/:id", userCtrl.Get)

	auth.GET("/logs", logCtrl.Search)
	auth.POST("/log/refresh", logCtrl.Refresh)

	auth.GET("/sales", saleCtrl.Search)
	auth.GET("/daily-sales", saleCtrl.DailySales)
	auth.GET("/monthly-sales", saleCtrl.MonthlySales)

	auth.GET("/vendo-status-history", statusCtrl.Search)

	auth.GET("/vendo-machines", vendoCtrl.All)
	auth.GET("/vendo-machines/:id/status", vendoCtrl.Status)
	auth.GET("/vendo-machines/:id", vendoCtrl.Get)
	auth.POST("/vendo-machines", vendoCtrl.Store)
	auth.DELETE("/vendo-machines/:id", vendoCtrl.Delete)
	auth.POST("/vendo-machines/:id/withdraw-current-sales", vendoCtrl.Withdraw)
	auth.POST("/vendo-machines/:id/set-status", vendoCtrl.SetStatus)

	auth.GET("/withdrawals", withdrawalCtrl.Search)

	return router
}

// corsMiddleware adds CORS headers matching the Python FastAPI configuration.
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}
		if !allowed && strings.HasSuffix(origin, ".jaries.dev") {
			allowed = true
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

