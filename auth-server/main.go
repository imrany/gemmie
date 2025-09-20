package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	v1 "github.com/imrany/gemmie/auth-server/internal/handlers"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	r := mux.NewRouter()

	// Routes
	r.HandleFunc("/api/register", v1.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/login", v1.LoginHandler).Methods("POST")
	r.HandleFunc("/api/sync", v1.SyncHandler).Methods("GET", "POST")
	r.HandleFunc("/api/health", v1.HealthHandler).Methods("GET")

	// CORS middleware
	corsOptions := handlers.AllowedOrigins([]string{"*"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "X-User-ID", "Authorization"})

	// Load .env if present
	if err := godotenv.Load(); err != nil {
		slog.Warn("Using default config, .env file not found")
	}

	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8081")
	
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	slog.Info("Logging initialized")

	// Start server
	port := viper.GetString("PORT")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(corsOptions, corsMethods, corsHeaders)(r)))
}