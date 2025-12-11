package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "github.com/imrany/gemmie/gemmie-server/internal/handlers"
	"github.com/imrany/gemmie/gemmie-server/internal/handlers/public"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/imrany/whats-email/pkg/mailer"
	"github.com/imrany/whats-email/pkg/whatsapp"

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

func runServer() {
	port := viper.GetInt("PORT")

	DSN := viper.GetString("DSN")
	whatsappDBPath := viper.GetString("WHATSAPP_DB_PATH")

	slog.Info("Starting server", "port", port, "DSN", DSN)

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

	// Initialize WhatsApp client
	slog.Info("Initializing WhatsApp client...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := whatsapp.Init(ctx, &whatsappDBPath); err != nil {
		slog.Error("Error initializing WhatsApp client", "error", err.Error())
		slog.Warn("Server will start without WhatsApp integration")
		// Continue running server even if WhatsApp fails to initialize
	} else {
		slog.Info("WhatsApp client initialized successfully")
	}

	// Configure email scheduler
	schedulerConfig := v1.EmailSchedulerConfig{
		SMTPConfig:      smtpConfig,
		SendInterval:    7 * 24 * time.Hour,    // Send every 7 days (recommended)
		EnableScheduler: smtpConfig.Host != "", // Only enable if SMTP is configured
	}

	// Start the email scheduler in background
	v1.StartEmailScheduler(schedulerConfig)

	// migrations run automatically
	if err := store.InitStorage(DSN); err != nil {
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

	// Arcade routes
	r.HandleFunc("/api/arcades", v1.CreateArcadeHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/arcades", v1.GetArcadesHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/arcades/{id}", v1.GetArcadeHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/arcades/{id}", v1.UpdateArcadeHandler).Methods(http.MethodPut)
	r.HandleFunc("/api/arcades/{id}", v1.DeleteArcadeHandler).Methods(http.MethodDelete)

	// Message routes
	r.HandleFunc("/api/chats/{id}/messages", v1.CreateMessageHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/chats/{id}/messages", v1.UpdateMessageHandler).Methods(http.MethodPut)
	r.HandleFunc("/api/messages/{id}", v1.DeleteMessageHandler).Methods(http.MethodDelete)

	// Errors Handler - stores user errors for later support and fix
	r.HandleFunc("/api/errors", v1.ErrorsHandler).Methods(http.MethodPost, http.MethodGet, http.MethodDelete)
	r.HandleFunc("/api/errors/{id}", v1.ErrorHandler).Methods(http.MethodDelete, http.MethodGet, http.MethodPut)

	// Payment routes
	r.HandleFunc("/api/payments/stk", v1.SendSTKHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/transactions", v1.GetTransactionsHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/transactions/{external_reference}", v1.GetTransactionByRefHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/callback", v1.StoreTransactionHandler).Methods(http.MethodPost)

	// Push notification endpoints
	r.HandleFunc("/api/push/subscribe", v1.SubscribeToPushNotificationHandler).Methods("POST")
	r.HandleFunc("/api/push/unsubscribe", v1.UnsubscribeToPushNotificationHandler).Methods("POST")
	r.HandleFunc("/api/push/send", v1.SendPushNotificationHandler).Methods("POST")
	r.HandleFunc("/api/push/subscriptions", v1.GetUserSubscriptionsHandler).Methods("GET")
	r.HandleFunc("/api/push/verify-subscription", v1.VerifySubscriptionHandler).Methods("POST")

	// genai routes
	r.HandleFunc("/api/genai", v1.GenerateAIResponseHandler).Methods(http.MethodPost)

	// Email management routes
	r.HandleFunc("/unsubscribe", v1.UnsubscribeHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/resubscribe", v1.ResubscribeHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/email-subscription", v1.UpdateEmailSubscriptionHandler).Methods(http.MethodPut)

	// Email verification routes
	r.HandleFunc("/api/send-verification", func(w http.ResponseWriter, r *http.Request) {
		v1.SendVerificationEmailHandler(w, r, smtpConfig)
	}).Methods(http.MethodPost)

	r.HandleFunc("/api/verify-email", v1.VerifyEmailHandler).Methods(http.MethodGet, http.MethodPost)

	r.HandleFunc("/api/ocr/upload", v1.OCRUploadHandler).Methods(http.MethodPost)

	// Email sending route (for Supabase Edge Function)
	r.HandleFunc("/api/email/send", func(w http.ResponseWriter, r *http.Request) {
		public.SendEmailHandler(w, r, smtpConfig)
	}).Methods(http.MethodPost)
	// Whatsapp message sending (for Supabase Edge Function)
	r.HandleFunc("/api/whatsapp/send", public.WhatsAppHandler).Methods(http.MethodPost)

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
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
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

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server exited cleanly")
	}
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
	rootCmd := &cobra.Command{
		Use:   "gemmie-server",
		Short: "Auth Server with sync functionality",
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	generateVapidCmd := &cobra.Command{
		Use:   "generate-vapid",
		Short: "Generate VAPID keys",
		Run: func(cmd *cobra.Command, args []string) {
			v1.GenerateVAPIDKeys()
		},
	}

	rootCmd.AddCommand(generateVapidCmd)

	envBindings := map[string]string{
		"port":               "PORT",
		"dsn":                "DSN",
		"payhero-username":   "PAYHERO_USERNAME",
		"payhero-password":   "PAYHERO_PASSWORD",
		"payhero-channel-id": "PAYHERO_CHANNEL_ID",
		"callback-url":       "CALLBACK_URL",
		"smtp-host":          "SMTP_HOST",
		"smtp-port":          "SMTP_PORT",
		"smtp-username":      "SMTP_USERNAME",
		"smtp-password":      "SMTP_PASSWORD",
		"smtp-email":         "SMTP_EMAIL",
		"whatsapp-db-path":   "WHATSAPP_DB_PATH",
		"api-key":            "API_KEY",
		"model":              "MODEL",
		"log-level":          "LOG_LEVEL",
		"vapid-public-key":   "VAPID_PUBLIC_KEY",
		"vapid-private-key":  "VAPID_PRIVATE_KEY",
		"vapid-email":        "VAPID_EMAIL",
	}

	rootCmd.PersistentFlags().Int("port", 8080, "Port to listen on (env: PORT)")
	rootCmd.PersistentFlags().String("dsn", "", "DSN (env: DSN)")
	rootCmd.PersistentFlags().String("payhero-username", "", "PayHero Username (env: PAYHERO_USERNAME)")
	rootCmd.PersistentFlags().String("payhero-password", "", "PayHero Password (env: PAYHERO_PASSWORD)")
	rootCmd.PersistentFlags().String("payhero-channel-id", "", "PayHero channel ID (env: PAYHERO_CHANNEL_ID)")
	rootCmd.PersistentFlags().String("callback-url", "", "Callback URL for PayHero (env: CALLBACK_URL)")
	rootCmd.PersistentFlags().String("smtp-host", "", "SMTP HOST (env: SMTP_HOST)")
	rootCmd.PersistentFlags().Int("smtp-port", 587, "SMTP PORT (env: SMTP_PORT)")
	rootCmd.PersistentFlags().String("smtp-username", "", "SMTP Username (env: SMTP_USERNAME)")
	rootCmd.PersistentFlags().String("smtp-password", "", "SMTP Password (env: SMTP_PASSWORD)")
	rootCmd.PersistentFlags().String("smtp-email", "", "SMTP Email (env: SMTP_EMAIL)")
	rootCmd.PersistentFlags().String("whatsapp-db-path", "", "WhatsApp Database Path (env: WHATSAPP_DB_PATH)")
	rootCmd.PersistentFlags().String("api-key", "", "API Key (env: API_KEY)")
	rootCmd.PersistentFlags().String("model", "", "Model (env: MODEL)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error) (env: LOG_LEVEL)")
	rootCmd.PersistentFlags().String("vapid-public-key", "", "VAPID Public Key (env: VAPID_PUBLIC_KEY)")
	rootCmd.PersistentFlags().String("vapid-private-key", "", "VAPID Private Key (env: VAPID_PRIVATE_KEY)")
	rootCmd.PersistentFlags().String("vapid-email", "", "VAPID Email (env: VAPID_EMAIL)")

	for key, env := range envBindings {
		if err := viper.BindPFlag(env, rootCmd.PersistentFlags().Lookup(key)); err != nil {
			panic(fmt.Errorf("failed to bind flag '%s': %s", key, err.Error()))
		}
	}
	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute command", "error", err)
		os.Exit(1)
	}
}
