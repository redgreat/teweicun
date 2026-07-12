/**
 * 功能：主程序入口
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/config"
	"github.com/redgreat/teweicun/internal/handler"
	"github.com/redgreat/teweicun/internal/middleware"
	"github.com/redgreat/teweicun/pkg/cache"
	"github.com/redgreat/teweicun/pkg/database"
	"github.com/redgreat/teweicun/pkg/logger"
	"github.com/redgreat/teweicun/pkg/storage"
	"go.uber.org/zap"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "c", "", "path to config file")
	flag.Parse()

	// 1. Load Configuration
	if err := config.LoadConfig(cfgPath); err != nil {
		log.Fatalf("Fatal error loading config: %v", err)
	}
	timezone := config.GlobalConfig.Database.Timezone
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		if timezone == "Asia/Shanghai" {
			// 某些精简镜像未包含 tzdata，此时退化为固定东八区避免服务启动失败。
			loc = time.FixedZone("CST", 8*3600)
			log.Printf("Timezone %s unavailable, fallback to fixed UTC+8: %v", timezone, err)
		} else {
			log.Fatalf("Fatal error loading timezone %s: %v", timezone, err)
		}
	}
	time.Local = loc

	// 2. Initialize Logger
	isDev := config.GlobalConfig.Server.Mode == "debug"
	if err := logger.InitLogger(config.GlobalConfig.Log.Level, isDev, config.GlobalConfig.Log.Filename); err != nil {
		log.Fatalf("Fatal error initializing logger: %v", err)
	}
	defer logger.Sync()

	logger.Log.Info("Starting TeWeiCun Server...", zap.String("timezone", timezone))

	// 3. Initialize Database
	if err := database.InitPostgres(config.GlobalConfig.Database); err != nil {
		logger.Log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.ClosePostgres()

	// 4. Initialize Redis
	if err := cache.InitRedis(config.GlobalConfig.Redis); err != nil {
		logger.Log.Fatal("Failed to initialize redis", zap.Error(err))
	}
	defer cache.CloseRedis()

	// 5. Initialize Storage
	if config.GlobalConfig.Storage.Type == "local" {
		s, err := storage.NewLocalStorage(config.GlobalConfig.Storage.LocalPath)
		if err != nil {
			logger.Log.Fatal("Failed to initialize local storage", zap.Error(err))
		}
		storage.GlobalStorage = s
	}

	// 6. Setup Gin Router
	if !isDev {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.CustomRecovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	// Basic Health Check Route
	r.GET("/api/v1/health", func(c *gin.Context) {
		// Verify DB connection
		dbStatus := "ok"
		if err := database.Pool.Ping(c.Request.Context()); err != nil {
			dbStatus = "error: " + err.Error()
		}

		// Verify Redis connection
		redisStatus := "ok"
		if _, err := cache.Client.Ping(c.Request.Context()).Result(); err != nil {
			redisStatus = "error: " + err.Error()
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
			"dependencies": gin.H{
				"postgres": dbStatus,
				"redis":    redisStatus,
			},
		})
	})

	// Static routes for local storage
	if config.GlobalConfig.Storage.Type == "local" {
		r.Static("/uploads", config.GlobalConfig.Storage.LocalPath)
	}

	// Add auth routes
	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/auth/login", handler.Login)

		// Protected routes example
		protected := apiV1.Group("")
		protected.Use(middleware.AuthJWT())
		protected.Use(func(c *gin.Context) {
			values := c.Request.URL.Query()
			if values.Get("page") == "" {
				values.Set("page", "1")
			}
			if values.Get("page_size") == "" {
				if limit := values.Get("limit"); limit != "" {
					values.Set("page_size", limit)
				} else {
					values.Set("page_size", "10")
				}
			}
			c.Request.URL.RawQuery = values.Encode()
			c.Next()
		})
		{
			protected.GET("/auth/me", func(c *gin.Context) {
				userID, _ := middleware.GetUserID(c)
				username, _ := c.Get("username")
				c.JSON(http.StatusOK, gin.H{
					"user_id":  userID,
					"username": username,
				})
			})

			// Dictionary Routes
			protected.GET("/system/dict-types", handler.ListDictTypes)
			protected.POST("/system/dict-types", handler.CreateDictType)
			protected.PUT("/system/dict-types/:id", handler.UpdateDictType)
			protected.DELETE("/system/dict-types/:id", handler.DeleteDictType)
			protected.GET("/system/dict/:type/data", handler.ListDictData)
			protected.POST("/system/dict-data", handler.CreateDictData)
			protected.PUT("/system/dict-data/:id", handler.UpdateDictData)
			protected.DELETE("/system/dict-data/:id", handler.DeleteDictData)

			// Material Category Routes
			protected.GET("/base/categories/tree", handler.GetCategoryTree)
			protected.POST("/base/categories", handler.CreateCategory)
			protected.PUT("/base/categories/:id", handler.UpdateCategory)
			protected.DELETE("/base/categories/:id", handler.DeleteCategory)

			// Material Routes
			protected.GET("/base/materials", handler.ListMaterials)
			protected.GET("/base/materials/:id", handler.GetMaterial)
			protected.POST("/base/materials", handler.CreateMaterial)
			protected.PUT("/base/materials/:id", handler.UpdateMaterial)

			// Supplier Routes
			protected.GET("/base/suppliers", handler.ListSuppliers)
			protected.POST("/base/suppliers", handler.CreateSupplier)
			protected.PUT("/base/suppliers/:id", handler.UpdateSupplier)
			protected.DELETE("/base/suppliers/:id", handler.DeleteSupplier)

			// Warehouse Routes
			protected.GET("/base/warehouses", handler.ListWarehouses)
			protected.POST("/base/warehouses", handler.CreateWarehouse)
			protected.PUT("/base/warehouses/:id", handler.UpdateWarehouse)
			protected.DELETE("/base/warehouses/:id", handler.DeleteWarehouse)

			// Customer Routes
			protected.GET("/base/customers", handler.ListCustomers)
			protected.POST("/base/customers", handler.CreateCustomer)
			protected.PUT("/base/customers/:id", handler.UpdateCustomer)
			protected.DELETE("/base/customers/:id", handler.DeleteCustomer)
			protected.GET("/base/partners/dropdown", handler.ListPartnerDropdown)

			// Certificate Routes
			protected.GET("/base/certificates", handler.ListCertificates)
			protected.POST("/base/certificates", handler.CreateCertificate)
			protected.PUT("/base/certificates/:id", handler.UpdateCertificate)
			protected.DELETE("/base/certificates/:id", handler.DeleteCertificate)

			// Purchase Request Routes (TODO: implement handlers)
			// protected.GET("/purchase/requests", handler.ListPurchaseRequests)
			// protected.GET("/purchase/requests/:id", handler.GetPurchaseRequestDetail)
			// protected.POST("/purchase/requests", handler.CreatePurchaseRequest)
			// protected.POST("/purchase/requests/:id/approve", handler.ApprovePurchaseRequest)

			// Purchase Order Routes
			protected.GET("/purchase/orders", handler.ListPurchaseOrders)
			protected.GET("/purchase/orders/:id", handler.GetPurchaseOrder)
			protected.POST("/purchase/orders", handler.CreatePurchaseOrder)
			protected.PUT("/purchase/orders/:id", handler.UpdatePurchaseOrder)
			protected.DELETE("/purchase/orders/:id", handler.DeletePurchaseOrder)
			protected.POST("/purchase/orders/:id/confirm", handler.ConfirmPurchaseOrder)

			// Fund Routes
			protected.GET("/fund/payments", handler.ListFundPayments)
			protected.GET("/fund/payment-sources", handler.ListFundPaymentSources)
			protected.GET("/fund/payments/:id", handler.GetFundPayment)
			protected.POST("/fund/payments", handler.CreateFundPayment)

			protected.GET("/fund/collections", handler.ListFundCollections)
			protected.GET("/fund/collection-sources", handler.ListFundCollectionSources)
			protected.GET("/fund/collections/:id", handler.GetFundCollection)
			protected.POST("/fund/collections", handler.CreateFundCollection)

			// Stock In Routes
			protected.GET("/stock-in", handler.ListStockIns)
			protected.GET("/stock-in/:id", handler.GetStockIn)
			protected.GET("/stock-in/:id/confirm-logs", handler.ListStockInConfirmLogs)
			protected.POST("/stock-in", handler.CreateStockIn)
			protected.PUT("/stock-in/:id", handler.UpdateStockIn)
			protected.POST("/stock-in/:id/confirm", handler.ConfirmStockIn)
			protected.POST("/stock-in/:id/confirm-reversal", handler.ConfirmReversalStockIn)
			protected.PUT("/stock-in-item/:id/serial-selections", handler.UpdateStockInItemSerialSelections)

			// Stock Out Routes
			protected.GET("/stock-out", handler.ListStockOuts)
			protected.GET("/stock-out/:id", handler.GetStockOutDetail)
			protected.POST("/stock-out", handler.CreateStockOut)
			protected.POST("/stock-out/:id/confirm", handler.ConfirmStockOut)
			protected.PUT("/stock-out/:id/serial-selections", handler.UpdateStockOutSerialSelections)
			protected.PUT("/stock-out-item/:id/serial-selections", handler.UpdateStockOutItemSerialSelections)

			// Consumption Order Routes (领料订单)
			protected.GET("/consumption/orders", handler.ListConsumptionOrders)
			protected.GET("/consumption/orders/:id", handler.GetConsumptionOrderDetail)
			protected.POST("/consumption/orders", handler.CreateConsumptionOrder)
			protected.PUT("/consumption/orders/:id", handler.UpdateConsumptionOrder)
			protected.POST("/consumption/orders/:id/confirm", handler.ConfirmConsumptionOrder)
			protected.DELETE("/consumption/orders/:id", handler.DeleteConsumptionOrder)

			// Reversal Order Routes (退料订单)
			protected.GET("/reversal/orders", handler.ListReversalOrders)
			protected.GET("/reversal/orders/:id", handler.GetReversalOrderDetail)
			protected.POST("/reversal/orders", handler.CreateReversalOrder)
			protected.PUT("/reversal/orders/:id", handler.UpdateReversalOrder)
			protected.POST("/reversal/orders/:id/confirm", handler.ConfirmReversalOrder)
			protected.DELETE("/reversal/orders/:id", handler.DeleteReversalOrder)

			// Production Routes (生产单/生产退货单)
			protected.GET("/production/orders", handler.ListProductionOrders)
			// 下拉列表必须在 :id 之前注册
			protected.GET("/production/orders/dropdown", handler.ListProductionOrdersForDropdown)
			protected.GET("/production/orders/:id", handler.GetProductionOrderDetail)
			protected.PUT("/production/orders/:id", handler.UpdateProductionOrder)
			protected.GET("/production/orders/:id/consumption-orders", handler.ListConsumptionOrdersByProduction)
			protected.GET("/production/orders/:id/reversal-orders", handler.ListReversalOrdersByProduction)
			protected.GET("/production/returns", handler.ListProductionReturnOrders)
			// 下拉列表必须在 :id 之前注册
			protected.POST("/production/returns", handler.CreateProductionReturnOrder)
			protected.GET("/production/returns/dropdown", handler.ListProductionReturnOrdersForDropdown)
			protected.GET("/production/returns/:id", handler.GetProductionReturnOrderDetail)
			protected.PUT("/production/returns/:id", handler.UpdateProductionReturnOrder)
			protected.GET("/production/returns/:id/consumption-orders", handler.ListConsumptionOrdersByProductionReturn)
			protected.GET("/production/returns/:id/reversal-orders", handler.ListReversalOrdersByProductionReturn)

			// Stock Transfer Routes
			protected.GET("/stock-transfers", handler.ListStockTransfers)
			protected.GET("/stock-transfers/:id", handler.GetStockTransferDetail)
			protected.POST("/stock-transfers", handler.CreateStockTransfer)
			protected.POST("/stock-transfers/:id/confirm-out", handler.ConfirmTransferOut)
			protected.POST("/stock-transfers/:id/confirm-in", handler.ConfirmTransferIn)

			// Inventory Check Routes
			protected.GET("/inventory-checks", handler.ListInventoryChecks)
			protected.GET("/inventory-checks/:id", handler.GetInventoryCheckDetail)
			protected.POST("/inventory-checks", handler.CreateInventoryCheck)
			protected.POST("/inventory-checks/:id/confirm", handler.ConfirmInventoryCheck)

			// Inventory Alert Routes
			protected.GET("/inventory/alerts", handler.ListInventoryAlerts)
			protected.POST("/inventory/alerts/check", handler.CheckInventoryAlerts)

			// Inventory Routes
			protected.GET("/inventory/detail", handler.ListInventoryDetail)
			protected.GET("/inventory/summary", handler.ListInventorySummary)
			protected.GET("/inventory/available", handler.ListInventoryAvailable)
			protected.GET("/inventory/issued", handler.ListInventoryIssued)
			protected.GET("/inventory/material-ledger", handler.ListInventoryMaterialLedger)
			protected.GET("/inventory/material-ledger/serials", handler.ListInventoryMaterialLedgerSerials)
			protected.GET("/inventory/material-ledger/export", handler.ExportInventoryMaterialLedger)
			protected.GET("/inventory/sku-ledger", handler.ListInventorySKULedger)
			protected.GET("/inventory/sku-ledger/serials", handler.ListInventorySKUSerials)
			protected.GET("/inventory/sku-ledger/export", handler.ExportInventorySKULedger)

			// Sales Order Routes
			protected.GET("/sales/orders", handler.ListSalesOrders)
			protected.GET("/sales/orders/:id", handler.GetSalesOrderDetail)
			protected.POST("/sales/orders", handler.CreateSalesOrder)
			protected.PUT("/sales/orders/:id", handler.UpdateSalesOrder)
			protected.POST("/sales/orders/:id/confirm", handler.ConfirmSalesOrder)
			protected.POST("/sales/orders/:id/cancel", handler.CancelSalesOrder)
			protected.POST("/sales/orders/:id/ship", handler.ShipSalesOrder)

			// Return Order Routes
			protected.GET("/returns", handler.ListReturnOrders)
			protected.GET("/returns/:id", handler.GetReturnOrderDetail)
			protected.POST("/returns", handler.CreateReturnOrder)
			protected.PUT("/returns/:id", handler.UpdateReturnOrder)
			protected.PUT("/returns/:id/sales", handler.UpdateSalesReturnOrder)
			protected.POST("/returns/:id/confirm", handler.ConfirmReturnOrder)
			protected.DELETE("/returns/:id", handler.DeleteReturnOrder)

			// Trace Routes
			protected.GET("/trace/forward", handler.TraceForward)
			protected.GET("/trace/backward", handler.TraceBackward)
			protected.GET("/trace/material/serial", handler.TraceMaterialBySerial)
			protected.GET("/dashboard/bigscreen", handler.GetBigscreenDashboard)

			// Material Serial Code Routes
			protected.GET("/serial-codes/stock-in-item/:id", handler.GetMaterialSerialCodesByStockInItem)
			protected.GET("/serial-codes/stock-in-item/:id/available-issued", handler.GetAvailableIssuedMaterialSerialCodesByStockInItem)
			protected.GET("/serial-codes/stock-out-item/:id", handler.GetMaterialSerialCodesByStockOutItem)
			protected.GET("/serial-codes/stock-out-item/:id/available", handler.GetAvailableMaterialSerialCodesByStockOutItem)

			// Legacy Serial Code Route Aliases
			protected.GET("/sku-serial/stock-in-item/:id", handler.GetMaterialSerialCodesByStockInItem)
			protected.GET("/sku-serial/stock-in-item/:id/available-issued", handler.GetAvailableIssuedMaterialSerialCodesByStockInItem)
			protected.GET("/sku-serial/stock-out-item/:id", handler.GetMaterialSerialCodesByStockOutItem)
			protected.GET("/sku-serial/stock-out-item/:id/available", handler.GetAvailableMaterialSerialCodesByStockOutItem)

			// Report Routes
			protected.GET("/reports/stock-in", handler.ReportStockInSummary)
			protected.GET("/reports/stock-out", handler.ReportStockOutSummary)
			protected.GET("/reports/inventory", handler.ReportInventoryStatus)
			protected.GET("/reports/balance", handler.ReportInventoryBalance)
			protected.GET("/reports/turnover", handler.ReportInventoryTurnover)
			protected.GET("/reports/reconciliation/customers", handler.ReportCustomerReconciliationSummary)
			protected.GET("/reports/reconciliation/suppliers", handler.ReportSupplierReconciliationSummary)
			protected.GET("/reports/profit", handler.ReportProfit)

			// Notification Routes
			protected.GET("/notifications", handler.ListNotifications)
			protected.POST("/notifications/:id/read", handler.MarkNotificationRead)

			// Upload Routes
			protected.POST("/system/upload", handler.UploadFile)

			// System User Management Routes
			protected.GET("/system/users", handler.ListUsers)
			protected.POST("/system/users", handler.CreateUser)
			protected.PUT("/system/users/:id", handler.UpdateUser)
			protected.DELETE("/system/users/:id", handler.DeleteUser)
			protected.PUT("/system/users/:id/password", handler.UpdatePassword)
			protected.POST("/system/users/:id/roles", handler.AssignUserRoles)

			// System Role Management Routes
			protected.GET("/system/roles", handler.ListRoles)
			protected.GET("/system/roles/:id", handler.GetRole)
			protected.POST("/system/roles", handler.CreateRole)
			protected.PUT("/system/roles/:id", handler.UpdateRole)
			protected.DELETE("/system/roles/:id", handler.DeleteRole)
			protected.POST("/system/roles/:id/permissions", handler.SetRolePermissions)

			// System Permission Routes
			protected.GET("/system/permissions/tree", handler.GetPermissionTree)
		}
	}

	// Serve frontend static files built by SvelteKit adapter-static.
	frontendDistDir := "./frontend-dist"
	r.StaticFS("/_app", gin.Dir(filepath.Join(frontendDistDir, "_app"), false))
	r.StaticFile("/favicon.png", filepath.Join(frontendDistDir, "favicon.png"))
	r.StaticFile("/robots.txt", filepath.Join(frontendDistDir, "robots.txt"))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/uploads/") {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}
		c.File(filepath.Join(frontendDistDir, "index.html"))
	})

	// 6. Start HTTP Server with Graceful Shutdown
	addr := ":8080"
	if config.GlobalConfig.Server.Port != 0 {
		addr = ":" + time.Now().Format("0")[:0] // Just to avoid unused fmt, use string concat
		addr = fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Log.Info("Server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Listen error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}
