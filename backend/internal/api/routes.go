package api

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/namespaces", GetNamespaces)
		api.GET("/resources/types", GetResourceTypes)
		api.GET("/resources", ListResources)

		api.GET("/pod", GetPodDetails)

		// ✅ NEW: get pod containers list for logs dropdown
		api.GET("/pod/containers", GetPodContainers)

		api.GET("/pod/events", GetPodEvents)
		api.GET("/pod/metrics", GetPodMetrics)

		// ✅ Logs streaming (supports container)
		api.GET("/logs/stream", StreamPodLogsSSE)
	}
}
