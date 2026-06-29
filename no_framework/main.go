package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/carloscfgos1980/cashier-app/internal/database"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

// apiConfig holds the dependencies for the API handlers.
type apiConfig struct {
	db   *database.Queries
	port string
}

func main() {
	// Load environment variables from .env file
	godotenv.Load()
	// Get configuration from environment variables
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	// Get the port from environment variables, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}
	// Connect to the database
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer dbConn.Close()

	// database queries variable
	dbQueries := database.New(dbConn)
	// variable for the apiConfig struct
	apiCfg := apiConfig{
		db:   dbQueries,
		port: port,
	}
	// Set up the HTTP server and routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", apiCfg.handlerHealth)
	mux.HandleFunc("GET /api/bills", apiCfg.handlerBillsGet)
	mux.HandleFunc("POST /api/bills", apiCfg.handlerBillsCreateUpdate)
	mux.HandleFunc("POST /api/change", apiCfg.handlerGetChange)
	mux.Handle("/", http.FileServer(http.Dir("client")))

	// Start the HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server is running http://localhost:%s", port)
	// Listen and serve
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
