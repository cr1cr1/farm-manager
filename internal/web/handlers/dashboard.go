package handlers

import (
	"net/http"

	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterDashboardRoutes wires the protected dashboard and a demo fragment.
func RegisterDashboardRoutes(group *ghttp.RouterGroup) {
	d := &Dashboard{}
	// Protected root
	group.GET("/", d.DashboardGet)
	// Demo fragment endpoint to showcase hypermedia/DataStar swap.
	group.GET("/fragment/ping", d.PingFragment)
}

type Dashboard struct{}

// DashboardGet renders the blank dashboard shell with a placeholder area and demo button.
func (d *Dashboard) DashboardGet(r *ghttp.Request) {
	csrf := middleware.HiddenCsrfFieldValue(r)
	component := pages.DashboardPage(middleware.BasePath(), "Dashboard", csrf)
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = component.Render(r.GetCtx(), r.Response.Writer)
}

// PingFragment returns HTML containing an element with id=content for DataStar morphing.
func (d *Dashboard) PingFragment(r *ghttp.Request) {
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.WriteStatus(http.StatusOK)
	r.Response.Write(`<section id="content" class="card"><h3>Pong</h3><p>Hypermedia fragment loaded at ` + r.GetClientIp() + `.</p></section>`)
}
