package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/imrany/gemmie/gemmie-server/internal/handlers"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/imrany/gemmie/gemmie-server/internal/mailer"

	"log/slog"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setupLogging configures slog with proper output and level
func setupLogging() {
	// Determine log level from environment or default to info
	logLevel := slog.LevelInfo
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch levelStr {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}

	// Create handler with options
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time for better readability
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.000"))
				}
			}
			return a
		},
	})

	// Set as default logger
	slog.SetDefault(slog.New(handler))

	slog.Info("Logging configured", "level", logLevel.String())
}

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
	// Setup logging first so we can log everything
	setupLogging()

	// Load .env if present
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using defaults")
	} else {
		slog.Info(".env file loaded successfully")
	}

	// Root command with Cobra
	var rootCmd = &cobra.Command{
		Use:   "gemmie-server",
		Short: "Auth Server with sync functionality",
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	// Flags
	rootCmd.PersistentFlags().String("port", "8081", "Port to run the server on")
	rootCmd.PersistentFlags().String("data", "./gemmie_data.json", "Path to data file")
	rootCmd.PersistentFlags().String("PAYHERO_USERNAME", "", "PayHero username (env: PAYHERO_USERNAME)")
	rootCmd.PersistentFlags().String("PAYHERO_PASSWORD", "", "PayHero password (env: PAYHERO_PASSWORD)")
	rootCmd.PersistentFlags().String("PAYHERO_CHANNEL_ID", "", "PayHero channel ID (env: PAYHERO_CHANNEL_ID)")
	rootCmd.PersistentFlags().String("CALLBACK_URL", "", "Callback URL for PayHero (env: CALLBACK_URL)")
	rootCmd.PersistentFlags().String("SMTP_HOST", "", "SMTP HOST (env: SMTP_HOST)")
	rootCmd.PersistentFlags().Int("SMTP_PORT", 587, "SMTP PORT (env: SMTP_PORT)")
	rootCmd.PersistentFlags().String("SMTP_USERNAME", "", "SMTP Username (env: SMTP_USERNAME)")
	rootCmd.PersistentFlags().String("SMTP_PASSWORD", "", "SMTP Password (env: SMTP_PASSWORD)")
	rootCmd.PersistentFlags().String("SMTP_EMAIL", "", "SMTP Email (env: SMTP_EMAIL)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error) (env: LOG_LEVEL)")

	// Bind flags to viper
	viper.BindPFlag("PORT", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("DATA_FILE", rootCmd.PersistentFlags().Lookup("data"))
	viper.BindPFlag("PAYHERO_USERNAME", rootCmd.PersistentFlags().Lookup("PAYHERO_USERNAME"))
	viper.BindPFlag("PAYHERO_PASSWORD", rootCmd.PersistentFlags().Lookup("PAYHERO_PASSWORD"))
	viper.BindPFlag("PAYHERO_CHANNEL_ID", rootCmd.PersistentFlags().Lookup("PAYHERO_CHANNEL_ID"))
	viper.BindPFlag("CALLBACK_URL", rootCmd.PersistentFlags().Lookup("CALLBACK_URL"))
	viper.BindPFlag("SMTP_HOST", rootCmd.PersistentFlags().Lookup("SMTP_HOST"))
	viper.BindPFlag("SMTP_PORT", rootCmd.PersistentFlags().Lookup("SMTP_PORT"))
	viper.BindPFlag("SMTP_USERNAME", rootCmd.PersistentFlags().Lookup("SMTP_USERNAME"))
	viper.BindPFlag("SMTP_PASSWORD", rootCmd.PersistentFlags().Lookup("SMTP_PASSWORD"))
	viper.BindPFlag("SMTP_EMAIL", rootCmd.PersistentFlags().Lookup("SMTP_EMAIL"))
	viper.BindPFlag("LOG_LEVEL", rootCmd.PersistentFlags().Lookup("log-level"))

	// Bind env variables
	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute command", "error", err)
		os.Exit(1)
	}
}

func runServer() {
	port := viper.GetString("PORT")
	dataFile := viper.GetString("DATA_FILE")

	slog.Info("Starting server", "port", port, "data_file", dataFile)

	// Configure SMTP settings
	smtpConfig := mailer.SMTPConfig{
		Host:     viper.GetString("SMTP_HOST"),
		Port:     viper.GetInt("SMTP_PORT"),
		Username: viper.GetString("SMTP_USERNAME"),
		Password: viper.GetString("SMTP_PASSWORD"),
		Email:    viper.GetString("SMTP_EMAIL"),
	}

	// Validate SMTP configuration and log status
	if smtpConfig.Host == "" || smtpConfig.Username == "" || smtpConfig.Password == "" {
		slog.Warn("SMTP not fully configured, email features will be disabled", 
			"host_set", smtpConfig.Host != "",
			"username_set", smtpConfig.Username != "",
			"password_set", smtpConfig.Password != "",
		)
	} else {
		slog.Info("SMTP configured successfully", "host", smtpConfig.Host, "email", smtpConfig.Email)
	}

	// Configure email scheduler
	schedulerConfig := v1.EmailSchedulerConfig{
		SMTPConfig:      smtpConfig,
		SendInterval:    7 * 24 * time.Hour, // Send every 7 days (recommended)
		EnableScheduler: smtpConfig.Host != "", // Only enable if SMTP is configured
	}

	// Start the email scheduler in background
	v1.StartEmailScheduler(schedulerConfig)

	// Initialize storage
	store.InitStorage(dataFile);

	// Log storage status
	store.Storage.Mu.RLock()
	userCount := len(store.Storage.Users)
	transactionCount := len(store.Storage.Transactions)
	store.Storage.Mu.RUnlock()
	
	slog.Info("Storage initialized", "users", userCount, "transactions", transactionCount)

	// Router setup
	r := mux.NewRouter()
	
	// Auth routes
	r.HandleFunc("/api/register", v1.RegisterHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/login", v1.LoginHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/sync", v1.SyncHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/health", v1.HealthHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/delete_account", v1.DeleteAccountHandler).Methods(http.MethodDelete)
	r.HandleFunc("/api/profile", v1.ProfileHandler)

	// Payment routes
	r.HandleFunc("/api/payments/stk", v1.SendSTKHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/transactions", v1.GetTransactionsHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/transactions/{external_reference}", v1.GetTransactionByRefHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/callback", v1.StoreTransactionHandler).Methods(http.MethodPost)
	
	// Email management routes
	r.HandleFunc("/unsubscribe", v1.UnsubscribeHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/resubscribe", v1.ResubscribeHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/email-subscription", v1.UpdateEmailSubscriptionHandler).Methods(http.MethodPut)
	
	// Email verification routes
	r.HandleFunc("/api/send-verification", func(w http.ResponseWriter, r *http.Request) {
		v1.SendVerificationEmailHandler(w, r, smtpConfig)
	}).Methods(http.MethodPost)
	
	r.HandleFunc("/api/verify-email", v1.VerifyEmailHandler).Methods(http.MethodGet, http.MethodPost)
	
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
		slog.Info("Server starting", "port", port, "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	
	sig := <-quit
	slog.Info("Shutdown signal received", "signal", sig, "shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server exited cleanly")
	}
}