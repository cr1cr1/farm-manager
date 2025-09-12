package middleware

import (
	"bytes"

	"github.com/a-h/templ"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TemplRender renders the given templ.Component to a buffer and writes it to the response
// This avoids Writer closure by templ, that will cause loss of further response headers and weird bugs, like session cookies not being set.
func TemplRender(r *ghttp.Request, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(r.GetCtx(), &buf); err != nil {
		g.Log().Errorf(r.GetCtx(), "render template: %v", err)
		r.Response.WriteStatus(500, "Internal Server Error")
		return err
	}
	r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Response.Write(buf.Bytes())
	return nil
}
