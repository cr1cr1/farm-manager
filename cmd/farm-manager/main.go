package main

import (
	"os"

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
	s := g.Server()
	s.BindHandler("/healthz", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"version": version,
			"commit":  commit,
			"date":    date,
			"addr":    addrFromEnv(),
		})
	})
	s.SetAddr(addrFromEnv())
	s.Run()
}
