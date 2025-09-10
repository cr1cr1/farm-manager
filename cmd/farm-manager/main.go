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
	return ":8080"
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

	// Server.
	s := g.Server()
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
	handlers.RegisterDashboardRoutes(protected)

	s.Run()
}
