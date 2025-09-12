package handlers

import (
	"net/http"

	"github.com/cr1cr1/farm-manager/internal/data"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterDashboardRoutes wires the protected dashboard and a demo fragment.
func RegisterDashboardRoutes(group *ghttp.RouterGroup, repo data.UserRepo) {
	d := &Dashboard{Repo: repo}
	// Protected root
	group.GET("/", d.DashboardGet)
	// Demo fragment endpoint to showcase hypermedia/DataStar swap.
	group.GET("/fragment/ping", d.PingFragment)
}

type Dashboard struct {
	Repo data.UserRepo
}

// DashboardGet renders the blank dashboard shell with a placeholder area and demo button.
func (d *Dashboard) DashboardGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.DashboardPage(
			middleware.BasePath(),
			"Dashboard",
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
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
