// Package handlers - subscription handlers
package handlers

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/spf13/viper"
)

var (
	// Generate VAPID keys using: go run main.go generate-vapid
	VapidPublicKey  = viper.GetString("VAPID_PUBLIC_KEY")
	VapidPrivateKey = viper.GetString("VAPID_PRIVATE_KEY")
	VapidEmail      = viper.GetString("VAPID_EMAIL")
)

// GenerateVAPIDKeys - Generate VAPID keys (run once) - go run main.go generate-vapid
func GenerateVAPIDKeys() {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		slog.Error("Failed to generate VAPID keys", "Error", err)
		os.Exit(1)
	}
	slog.Info("GENERATED_VAPID_KEYS", "Public Key", publicKey)
	slog.Info("GENERATED_VAPID_KEYS", "Private Key", privateKey)
}

// SubscribeToPushNotificationHandler - Subscribes user to push notifications
func SubscribeToPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID required",
		})
		return
	}

	_, err := store.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	var sub store.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	userAgent := r.Header.Get("User-Agent")

	if err := store.SaveSubscription(r.Context(), userID, sub, userAgent); err != nil {
		log.Printf("Failed to save subscription: %v", err)
		http.Error(w, "Failed to save subscription", http.StatusInternalServerError)
		return
	}

	slog.Info("Subscription saved for user %s: %s", userID, sub.Endpoint)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Subscribed successfully",
	})
}

func UnsubscribeToPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID required",
		})
		return
	}

	_, err := store.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	var sub store.SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if err := store.DeleteSubscription(r.Context(), sub.Endpoint); err != nil {
		slog.Error("Failed to delete subscription", "Error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to unsubscribe",
		})
		return
	}

	slog.Info("Subscription deleted", "details: ", sub.Endpoint)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Unsubscribed successfully",
	})
}

func SendPushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID required",
		})
		return
	}

	_, err := store.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	var req store.SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get subscriptions
	subscriptions, err := store.GetSubscriptionsByUserIDs(r.Context(), req.UserIDs)
	if err != nil {
		slog.Error("Failed to get subscriptions", "Error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get subscriptions",
		})
		return
	}

	if len(subscriptions) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"success": 0,
			"failed":  0,
			"message": "No subscriptions found",
		})
		return
	}

	// Convert payload to JSON
	data, err := json.Marshal(req.Payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	successCount := 0
	failureCount := 0
	failedEndpoints := []string{}

	for _, sub := range subscriptions {
		resp, err := webpush.SendNotification(data, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.AuthKey,
				P256dh: sub.P256dhKey,
			},
		}, &webpush.Options{
			Subscriber:      VapidEmail,
			VAPIDPublicKey:  VapidPublicKey,
			VAPIDPrivateKey: VapidPrivateKey,
			TTL:             30,
		})

		if err != nil {
			log.Printf("Failed to send to %s: %v", sub.Endpoint, err)
			failureCount++
			failedEndpoints = append(failedEndpoints, sub.Endpoint)

			// Delete invalid subscriptions (410 Gone or 404 Not Found)
			if resp != nil && (resp.StatusCode == 410 || resp.StatusCode == 404) {
				store.DeleteSubscription(r.Context(), sub.Endpoint)
				slog.Info("Deleted invalid subscription", "Endpoint", sub.Endpoint)
			}
		} else {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				successCount++
			} else {
				slog.Info("Push failed", "Status", resp.StatusCode, "Endpoint", sub.Endpoint)
				failureCount++
				failedEndpoints = append(failedEndpoints, sub.Endpoint)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Push notifications sent",
		Data: map[string]any{
			"sent":             successCount,
			"failed":           failureCount,
			"failed_endpoints": failedEndpoints,
		},
	})
}

// GetUserSubscriptionsHandler -  Get user subscriptions
func GetUserSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID required",
		})
		return
	}

	_, err := store.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	subscriptions, err := store.GetSubscriptionsByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get subscriptions: %v", err)
		http.Error(w, "Failed to get subscriptions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Subscriptions retrieved successfully",
		Data:    subscriptions,
	})
}
