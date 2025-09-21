package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/imrany/gemmie/auth-server/internal/handlers"
	"github.com/imrany/gemmie/auth-server/store"

	"log/slog"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loggingResponseWriter captures status code for logging.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs method, path, query, status, duration, and remote IP for each request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", lrw.statusCode,
			"duration", duration,
			"remote", r.RemoteAddr,
		)
	})
}

func main() {
	// Load .env if present
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using defaults")
	}

	// Root command with Cobra
	var rootCmd = &cobra.Command{
		Use:   "auth-server",
		Short: "Auth Server with sync functionality",
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	// Flags
	rootCmd.PersistentFlags().String("port", "8081", "Port to run the server on")
	rootCmd.PersistentFlags().String("data", "./gemmie_data.json", "Path to data file")

	// Bind flags to viper
	viper.BindPFlag("PORT", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("DATA_FILE", rootCmd.PersistentFlags().Lookup("data"))

	// Bind env variables (PORT, DATA_FILE)
	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute command", "error", err)
		os.Exit(1)
	}
}

func runServer() {
	port := viper.GetString("PORT")
	dataFile := viper.GetString("DATA_FILE")

	// Initialize storage
	store.InitStorage(dataFile)

	// Router setup
	r := mux.NewRouter()
	r.HandleFunc("/api/register", v1.RegisterHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/login", v1.LoginHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/sync", v1.SyncHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/health", v1.HealthHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/delete_account", v1.DeleteAccountHandler).Methods(http.MethodDelete)


	// Handle CORS preflight
	r.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// CORS middleware
	corsOptions := handlers.AllowedOrigins([]string{"*"})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "X-User-ID", "Authorization"})
	corsCredentials := handlers.AllowCredentials()

	handler := handlers.CORS(corsOptions, corsMethods, corsHeaders, corsCredentials)(r)
	handler = loggingMiddleware(handler)

	// HTTP server
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("Server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server exited cleanly")
	}
}

