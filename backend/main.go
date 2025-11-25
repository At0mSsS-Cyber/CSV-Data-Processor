package main

import (
	"csv-processor/database"
	"csv-processor/handlers"
	"csv-processor/services"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Initialize services
	dbService := services.NewDBService()
	asyncProcessor := services.NewAsyncProcessor(dbService)

	// Initialize handlers
	h := handlers.NewHandler(dbService, asyncProcessor)

	// Setup router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/upload", h.HandleUpload).Methods("POST")
	router.HandleFunc("/api/files", h.HandleGetFiles).Methods("GET")
	router.HandleFunc("/api/files/{id}", h.HandleGetFile).Methods("GET")
	router.HandleFunc("/api/records", h.HandleGetRecords).Methods("GET")
	router.HandleFunc("/api/groups/records", h.HandleGetGroupRecords).Methods("GET")
	router.HandleFunc("/api/health", h.HandleHealth).Methods("GET")

	// CORS middleware
	router.Use(corsMiddleware)

	// Start server
	srv := &http.Server{
		Handler:      router,
		Addr:         ":8080",
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	log.Println("Server starting on port 8080...")
	log.Fatal(srv.ListenAndServe())
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
