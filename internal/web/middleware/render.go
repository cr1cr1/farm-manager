package middleware

import (
	"bytes"

	"github.com/a-h/templ"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func TemplRender(r *ghttp.Request, component templ.Component) error {
	var buf bytes.Buffer
	if err := component.Render(r.GetCtx(), &buf); err != nil {
		g.Log().Errorf(r.GetCtx(), "render template: %v", err)
		r.Response.WriteStatus(500, "Internal Server Error")
		return err
	}
	r.Response.Write(buf.Bytes())
	return nil
}
