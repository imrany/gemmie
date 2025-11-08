package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/imrany/gemmie/gemmie-server/internal/handlers"
	"github.com/imrany/gemmie/gemmie-server/pkg/mailer"
	"github.com/imrany/gemmie/gemmie-server/store"

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

	// Flags - UPDATED: Replace data file with database connection
	rootCmd.PersistentFlags().String("port", "8081", "Port to run the server on")
	rootCmd.PersistentFlags().String("db-host", "localhost", "Database host (env: DB_HOST)")
	rootCmd.PersistentFlags().String("db-port", "5432", "Database port (env: DB_PORT)")
	rootCmd.PersistentFlags().String("db-user", "", "Database user (env: DB_USER)")
	rootCmd.PersistentFlags().String("db-password", "", "Database password (env: DB_PASSWORD)")
	rootCmd.PersistentFlags().String("db-name", "gemmie", "Database name (env: DB_NAME)")
	rootCmd.PersistentFlags().String("db-sslmode", "disable", "Database SSL mode (env: DB_SSLMODE)")
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

	// Bind flags to viper - UPDATED: Database flags instead of data file
	viper.BindPFlag("PORT", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("DB_HOST", rootCmd.PersistentFlags().Lookup("db-host"))
	viper.BindPFlag("DB_PORT", rootCmd.PersistentFlags().Lookup("db-port"))
	viper.BindPFlag("DB_USER", rootCmd.PersistentFlags().Lookup("db-user"))
	viper.BindPFlag("DB_PASSWORD", rootCmd.PersistentFlags().Lookup("db-password"))
	viper.BindPFlag("DB_NAME", rootCmd.PersistentFlags().Lookup("db-name"))
	viper.BindPFlag("DB_SSLMODE", rootCmd.PersistentFlags().Lookup("db-sslmode"))
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

	dbHost := viper.GetString("DB_HOST")
	dbPort := viper.GetString("DB_PORT")
	dbUser := viper.GetString("DB_USER")
	dbPassword := viper.GetString("DB_PASSWORD")
	dbName := viper.GetString("DB_NAME")
	dbSSLMode := viper.GetString("DB_SSLMODE")

	// Build connection string
	connString := "host=" + dbHost +
		" port=" + dbPort +
		" user=" + dbUser +
		" password=" + dbPassword +
		" dbname=" + dbName +
		" sslmode=" + dbSSLMode

	slog.Info("Starting server", "port", port, "db_host", dbHost, "db_name", dbName)

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
		SendInterval:    7 * 24 * time.Hour,    // Send every 7 days (recommended)
		EnableScheduler: smtpConfig.Host != "", // Only enable if SMTP is configured
	}

	// Start the email scheduler in background
	v1.StartEmailScheduler(schedulerConfig)

	//migrations run automatically
	if err := store.InitStorage(connString); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Register cleanup on shutdown
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		} else {
			slog.Info("Database connection closed")
		}
	}()

	slog.Info("Database storage initialized successfully")

	// Router setup
	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/api/register", v1.RegisterHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/login", v1.LoginHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/sync", v1.SyncHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/health", v1.HealthHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/delete_account", v1.DeleteAccountHandler).Methods(http.MethodDelete)
	r.HandleFunc("/api/profile", v1.ProfileHandler)

	// Chat routes
	r.HandleFunc("/api/chats", v1.CreateChatHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/chats", v1.GetChatsHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/chats", v1.DeleteAllChatsHandler).Methods(http.MethodDelete)
	r.HandleFunc("/api/chats/{id}", v1.GetChatHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/chats/{id}", v1.UpdateChatHandler).Methods(http.MethodPut)
	r.HandleFunc("/api/chats/{id}", v1.DeleteChatHandler).Methods(http.MethodDelete)

	// Message routes
	r.HandleFunc("/api/chats/{id}/messages", v1.CreateMessageHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/chats/{id}/messages", v1.UpdateMessageHandler).Methods(http.MethodPut)
	r.HandleFunc("/api/messages/{id}", v1.DeleteMessageHandler).Methods(http.MethodDelete)

	// Errors Handler - stores user errors for later support and fix
	r.HandleFunc("/api/errors", v1.ErrorsHandler).Methods(http.MethodPost, http.MethodGet, http.MethodDelete)
	r.HandleFunc("/api/errors/:id", v1.ErrorHandler).Methods(http.MethodDelete, http.MethodGet, http.MethodPut)

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

	// Email sending route (for Supabase Edge Function)
	r.HandleFunc("/api/send-email", func(w http.ResponseWriter, r *http.Request) {
		v1.SendEmailHandler(w, r, smtpConfig)
	}).Methods(http.MethodPost)

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
	slog.Info("Shutdown signal received", "signal", sig, "shutting down gracefully...", "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server exited cleanly")
	}
}
