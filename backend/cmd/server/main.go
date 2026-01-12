package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"webk8s/internal/api"
	"webk8s/internal/web"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Serve UI + assets
	web.RegisterStatic(r)

	// API
	api.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	log.Printf("webk8s listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
