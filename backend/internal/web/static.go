package web

import "github.com/gin-gonic/gin"

func RegisterStatic(r *gin.Engine) {
	// Main UI
	r.StaticFile("/", "./ui/index.html")

	// ✅ Logs-only page (new tab)
	r.StaticFile("/logs.html", "./ui/logs.html")

	// UI assets
	r.Static("/assets", "./ui/assets")
}
