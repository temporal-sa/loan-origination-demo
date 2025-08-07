package main

import (
	"log"

	"loan-origination-system/internal/api"
	"loan-origination-system/pkg/temporal"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Temporal client
	temporalClient, err := temporal.NewClient()
	if err != nil {
		log.Fatal("Failed to create Temporal client:", err)
	}
	defer temporalClient.Close()

	// Setup Gin router
	router := gin.Default()

	// Enable CORS for development
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	api.SetupRoutes(router, temporalClient)

	log.Println("Server starting on :8082")
	if err := router.Run(":8082"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
