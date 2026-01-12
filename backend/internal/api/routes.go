package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Namespace and resource type endpoints
		api.GET("/namespaces", func(c *gin.Context) {
			log.Println("GET /api/namespaces")
			GetNamespaces(c)
		})

		api.GET("/resources/types", func(c *gin.Context) {
			log.Println("GET /api/resources/types")
			GetResourceTypes(c)
		})

		api.GET("/resources", func(c *gin.Context) {
			log.Printf("GET /api/resources?namespace=%s&type=%s", c.Query("namespace"), c.Query("type"))
			ListResources(c)
		})

		// Pod detail endpoints
		api.GET("/pod", func(c *gin.Context) {
			log.Printf("GET /api/pod?namespace=%s&pod=%s", c.Query("namespace"), c.Query("pod"))
			GetPodDetails(c)
		})

		api.GET("/pod/containers", func(c *gin.Context) {
			log.Printf("GET /api/pod/containers?namespace=%s&pod=%s", c.Query("namespace"), c.Query("pod"))
			GetPodContainers(c)
		})

		api.GET("/pod/events", func(c *gin.Context) {
			log.Printf("GET /api/pod/events?namespace=%s&pod=%s", c.Query("namespace"), c.Query("pod"))
			GetPodEvents(c)
		})

		api.GET("/pod/metrics", func(c *gin.Context) {
			log.Printf("GET /api/pod/metrics?namespace=%s&pod=%s", c.Query("namespace"), c.Query("pod"))
			GetPodMetrics(c)
		})

		// Node detail endpoints (NEW)
		api.GET("/node", func(c *gin.Context) {
			log.Printf("GET /api/node?node=%s", c.Query("node"))
			GetNodeDetails(c)
		})

		api.GET("/node/metrics", func(c *gin.Context) {
			log.Printf("GET /api/node/metrics?node=%s", c.Query("node"))
			GetNodeMetrics(c)
		})

		// Service detail endpoints (NEW)
		api.GET("/service", func(c *gin.Context) {
			log.Printf("GET /api/service?namespace=%s&service=%s", c.Query("namespace"), c.Query("service"))
			GetServiceDetails(c)
		})

		// ConfigMap detail endpoints (NEW)
		api.GET("/configmap", func(c *gin.Context) {
			log.Printf("GET /api/configmap?namespace=%s&configmap=%s", c.Query("namespace"), c.Query("configmap"))
			GetConfigMapDetails(c)
		})

		// Log streaming endpoint
		api.GET("/logs/stream", func(c *gin.Context) {
			log.Printf("GET /api/logs/stream?namespace=%s&pod=%s&container=%s",
				c.Query("namespace"), c.Query("pod"), c.Query("container"))
			StreamPodLogsSSE(c)
		})
	}

	log.Println("API routes registered successfully")
}
