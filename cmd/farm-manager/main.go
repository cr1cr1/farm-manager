package main

import (
	"context"
	"os"

	"github.com/cr1cr1/farm-manager/internal/data"
	appdb "github.com/cr1cr1/farm-manager/internal/db"
	"github.com/cr1cr1/farm-manager/internal/web/handlers"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gsession"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func addrFromEnv() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":3000"
}

func main() {
	ctx := context.Background()

	// Setup DB and run migrations.
	db, err := appdb.Open(ctx)
	if err != nil {
		panic(err)
	}
	defer appdb.Close(db)

	if err := appdb.Migrate(ctx, db); err != nil {
		panic(err)
	}

	// Repositories.
	userRepo := data.NewSQLiteUserRepo(db)
	barnRepo := &data.SQLiteBarnRepo{DB: db}
	feedTypeRepo := &data.SQLiteFeedTypeRepo{DB: db}
	staffRepo := &data.SQLiteStaffRepo{DB: db}
	flockRepo := &data.SQLiteFlockRepo{DB: db}
	feedingRecordRepo := &data.SQLiteFeedingRecordRepo{DB: db}
	healthCheckRepo := &data.SQLiteHealthCheckRepo{DB: db}
	mortalityRecordRepo := &data.SQLiteMortalityRecordRepo{DB: db}
	productionBatchRepo := &data.SQLiteProductionBatchRepo{DB: db}
	slaughterRecordRepo := &data.SQLiteSlaughterRecordRepo{DB: db}
	inventoryItemRepo := &data.SQLiteInventoryItemRepo{DB: db}
	customerRepo := &data.SQLiteCustomerRepo{DB: db}
	orderRepo := &data.SQLiteOrderRepo{DB: db}
	orderItemRepo := &data.SQLiteOrderItemRepo{DB: db}

	// Server.
	s := g.Server()
	// Session storage
	switch os.Getenv("APP_SESSION_STORE") {
	case "redis":
		s.SetSessionStorage(gsession.NewStorageRedis(g.Redis()))
	case "memory":
		s.SetSessionStorage(gsession.NewStorageMemory())
	default:
		// Default to file-based sessions
	}
	// Enable access and error logging via GoFrame logger (no sensitive data logged).
	s.SetAccessLogEnabled(true)
	s.SetErrorLogEnabled(true)
	s.SetAddr(addrFromEnv())

	// Health check (infra).
	s.BindHandler("/healthz", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"version": version,
			"commit":  commit,
			"date":    date,
			"addr":    addrFromEnv(),
		})
	})

	// Static assets under /public
	s.AddStaticPath("/public", "./public")

	// Global rate limit.
	s.Use(middleware.RateLimit())

	base := middleware.BasePath()

	// Public routes (login, logout). CSRF applied for POST.
	public := s.Group(base)
	public.Middleware(middleware.Csrf())
	handlers.RegisterAuthRoutes(public, userRepo)

	// Protected routes (dashboard and fragments).
	protected := s.Group(base)
	protected.Middleware(middleware.Csrf(), middleware.RequireAuth())

	// Create dashboard repos struct with all repositories
	dashboardRepos := &handlers.DashboardRepos{
		UserRepo:            userRepo,
		BarnRepo:            barnRepo,
		FeedTypeRepo:        feedTypeRepo,
		StaffRepo:           staffRepo,
		FlockRepo:           flockRepo,
		FeedingRecordRepo:   feedingRecordRepo,
		HealthCheckRepo:     healthCheckRepo,
		MortalityRecordRepo: mortalityRecordRepo,
		ProductionBatchRepo: productionBatchRepo,
		SlaughterRecordRepo: slaughterRecordRepo,
		InventoryItemRepo:   inventoryItemRepo,
		CustomerRepo:        customerRepo,
		OrderRepo:           orderRepo,
		OrderItemRepo:       orderItemRepo,
	}

	handlers.RegisterDashboardRoutes(protected, dashboardRepos)
	handlers.RegisterProfileRoutes(protected, userRepo)

	// Register individual domain management routes
	handlers.RegisterBarnRoutes(protected, barnRepo)
	handlers.RegisterFeedTypeRoutes(protected, feedTypeRepo)
	handlers.RegisterStaffRoutes(protected, staffRepo)
	handlers.RegisterFlockRoutes(protected, flockRepo, barnRepo, feedTypeRepo)
	handlers.RegisterFeedingRecordRoutes(protected, feedingRecordRepo, flockRepo, feedTypeRepo, staffRepo)
	handlers.RegisterHealthCheckRoutes(protected, healthCheckRepo, flockRepo, staffRepo)
	handlers.RegisterInventoryItemRoutes(protected, inventoryItemRepo)
	handlers.RegisterMortalityRecordRoutes(protected, mortalityRecordRepo, flockRepo)
	handlers.RegisterProductionBatchRoutes(protected, productionBatchRepo, flockRepo, staffRepo)
	handlers.RegisterSlaughterRecordRoutes(protected, slaughterRecordRepo, productionBatchRepo, staffRepo)
	handlers.RegisterCustomerRoutes(protected, customerRepo)
	handlers.RegisterOrderRoutes(protected, orderRepo, customerRepo)
	handlers.RegisterOrderItemRoutes(protected, orderItemRepo, orderRepo)

	s.Run()
}
