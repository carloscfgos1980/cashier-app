package main

import (
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/carloscfgos1980/cashier-app/internal/bills"
	"github.com/carloscfgos1980/cashier-app/internal/database"
	"github.com/carloscfgos1980/cashier-app/internal/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

type application struct {
	config       config
	db           *pgx.Conn
	requestCount atomic.Uint64
	startedAt    time.Time
}

// config holds the configuration for the application
type config struct {
	addr string
	db   dbConfig
}

// dbConfig holds the database configuration for the application
type dbConfig struct {
	dsn string
}

func (app *application) mount() http.Handler {
	if app.startedAt.IsZero() {
		app.startedAt = time.Now().UTC()
	}

	// create a new router
	r := chi.NewRouter()
	// set up middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.requestCount.Add(1)
			next.ServeHTTP(w, r)
		})
	})

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		health := make(map[string]string)
		health["status"] = "ok"
		health["version"] = "1.0.0"
		health["message"] = "Server is running"
		if err := json.WriteJSON(w, http.StatusOK, health); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		metrics := map[string]any{
			"uptime_seconds":      time.Since(app.startedAt).Seconds(),
			"requests_total":      app.requestCount.Load(),
			"goroutines":          runtime.NumGoroutine(),
			"memory_alloc_bytes":  mem.Alloc,
			"memory_heap_inuse":   mem.HeapInuse,
			"memory_sys_bytes":    mem.Sys,
			"gc_cycles_completed": mem.NumGC,
		}

		if err := json.WriteJSON(w, http.StatusOK, metrics); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	billsService := bills.NewService(database.New(app.db), app.db)
	billsHandler := bills.NewHandler(billsService)
	r.Get("/api/bills", billsHandler.GetBills)
	r.Post("/api/bills", billsHandler.BillsCreateUpdate)
	r.Post("/api/change", billsHandler.GetChange)

	// Serve the frontend client from the client/ directory.
	r.Handle("/*", http.FileServer(http.Dir("client")))
	return r
}

// run starts the HTTP server
func (app *application) run(h http.Handler) error {
	// create the HTTP server
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("Starting server on %s", app.config.addr)
	// start the server
	return srv.ListenAndServe()
}
