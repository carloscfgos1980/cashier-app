package main

import (
	"database/sql"
	"log"

	"github.com/carloscfgos1980/cashier-app/internal/database"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/carloscfgos1980/cashier-app/internal/config"
	"github.com/carloscfgos1980/cashier-app/internal/handlers"
)

func main() {
	// create a context

	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to the database
	dbConn, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer dbConn.Close()

	// Create a new database queries instance
	db := database.New(dbConn)

	cfg.DB = db

	// Initialize the Gin router
	var router *gin.Engine = gin.Default()

	// Set trusted proxies to nil to avoid warnings in Gin 1.7+
	router.SetTrustedProxies(nil)

	// Define a simple health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"version": "1.0",
			"message": "Cashier API is healthy",
		})
	})
	router.GET("/api/bills", handlers.GetBillsHandler(cfg))
	router.POST("/api/bills", handlers.BillsCreateUpdateHandler(cfg))
	router.POST("/api/change", handlers.GetChangeHandler(cfg))
	router.StaticFile("/", "./client/index.html")
	router.StaticFile("/index.html", "./client/index.html")
	router.StaticFile("/app.js", "./client/app.js")
	router.StaticFile("/styles.css", "./client/styles.css")

	// Start the server on the specified port
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
