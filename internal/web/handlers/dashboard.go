package handlers

import (
	"context"
	"net/http"

	"github.com/cr1cr1/farm-manager/internal/data"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/models"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterDashboardRoutes wires the protected dashboard and a demo fragment.
func RegisterDashboardRoutes(group *ghttp.RouterGroup, repos *DashboardRepos) {
	d := &Dashboard{Repos: repos}
	// Protected root
	group.GET("/", d.DashboardGet)
	// Demo fragment endpoint to showcase hypermedia/DataStar swap.
	group.GET("/fragment/ping", d.PingFragment)
}

type DashboardRepos struct {
	UserRepo            data.UserRepo
	BarnRepo            data.BarnRepo
	FeedTypeRepo        data.FeedTypeRepo
	StaffRepo           data.StaffRepo
	FlockRepo           data.FlockRepo
	FeedingRecordRepo   data.FeedingRecordRepo
	HealthCheckRepo     data.HealthCheckRepo
	MortalityRecordRepo data.MortalityRecordRepo
	ProductionBatchRepo data.ProductionBatchRepo
	SlaughterRecordRepo data.SlaughterRecordRepo
	InventoryItemRepo   data.InventoryItemRepo
	CustomerRepo        data.CustomerRepo
	OrderRepo           data.OrderRepo
	OrderItemRepo       data.OrderItemRepo
}

type Dashboard struct {
	Repos *DashboardRepos
}

// DashboardGet renders the dashboard with counts for all domain objects.
func (d *Dashboard) DashboardGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	ctx := context.Background()
	counts := &models.DashboardCounts{}

	// Get counts for all domain objects
	if count, err := d.Repos.BarnRepo.Count(ctx); err == nil {
		counts.Barns = count
	}
	if count, err := d.Repos.FeedTypeRepo.Count(ctx); err == nil {
		counts.FeedTypes = count
	}
	if count, err := d.Repos.StaffRepo.Count(ctx); err == nil {
		counts.Staff = count
	}
	if count, err := d.Repos.FlockRepo.Count(ctx); err == nil {
		counts.Flocks = count
	}
	if count, err := d.Repos.FeedingRecordRepo.Count(ctx); err == nil {
		counts.FeedingRecords = count
	}
	if count, err := d.Repos.HealthCheckRepo.Count(ctx); err == nil {
		counts.HealthChecks = count
	}
	if count, err := d.Repos.MortalityRecordRepo.Count(ctx); err == nil {
		counts.MortalityRecords = count
	}
	if count, err := d.Repos.ProductionBatchRepo.Count(ctx); err == nil {
		counts.ProductionBatches = count
	}
	if count, err := d.Repos.SlaughterRecordRepo.Count(ctx); err == nil {
		counts.SlaughterRecords = count
	}
	if count, err := d.Repos.InventoryItemRepo.Count(ctx); err == nil {
		counts.InventoryItems = count
	}
	if count, err := d.Repos.CustomerRepo.Count(ctx); err == nil {
		counts.Customers = count
	}
	if count, err := d.Repos.OrderRepo.Count(ctx); err == nil {
		counts.Orders = count
	}
	if count, err := d.Repos.OrderItemRepo.Count(ctx); err == nil {
		counts.OrderItems = count
	}

	_ = middleware.TemplRender(
		r,
		pages.DashboardPage(
			middleware.BasePath(),
			"Dashboard",
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			counts,
		),
	)
} // PingFragment returns HTML containing an element with id=content for DataStar morphing.
func (d *Dashboard) PingFragment(r *ghttp.Request) {
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.WriteStatus(http.StatusOK)
	r.Response.Write(`<div id="content" class="bg-card text-card-foreground rounded-xl border shadow-sm p-6">
		<h3 class="text-lg font-semibold mb-2">Pong! 🏓</h3>
		<p class="text-muted-foreground">Hypermedia fragment loaded at ` + r.GetClientIp() + `.</p>
		<p class="text-sm text-muted-foreground mt-2">This content was dynamically loaded via DataStar.</p>
	</div>`)
}
