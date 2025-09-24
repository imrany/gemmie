package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/spf13/viper"
)

// Request/Response types
type STKRequest struct {
	ExternalReference string `json:"external_reference"`
	Amount            int    `json:"amount"`
	PhoneNumber       string `json:"phone_number"`
}

type STKResponse struct {
	ExternalReference string      `json:"external_reference"`
	Data              interface{} `json:"data"`
}

type PayHeroResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Data access functions
func findTransactionByRef(ref string) (*store.Transaction, bool) {
	store.Storage.Mu.RLock()
	defer store.Storage.Mu.RUnlock()

	for _, transaction := range store.Storage.Transactions {
		if transaction.ExternalReference == ref {
			return &transaction, true
		}
	}
	return nil, false
}

func createTransaction(t store.Transaction) error {
	t.ID = encrypt.GenerateID("txn")
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	store.Storage.Mu.Lock()
	store.Storage.Transactions[t.ID] = t
	store.Storage.Mu.Unlock()

	store.SaveStorage()
	return nil
}

// --- helper: try to find user by email or username in storage ---
func findUserByEmailOrUsername(identifier string) (*store.User, string, bool) {
	// returns (user, userID, found)
	store.Storage.Mu.RLock()
	defer store.Storage.Mu.RUnlock()

	for id, u := range store.Storage.Users {
		if strings.EqualFold(u.Email, identifier) || strings.EqualFold(u.Username, identifier) {
			userCopy := u
			return &userCopy, id, true
		}
	}
	return nil, "", false
}

// --- SendSTKHandler: returns consistent store.Response with data object ---
func SendSTKHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var stkReq STKRequest
	if err := json.NewDecoder(r.Body).Decode(&stkReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Check if transaction already exists
	if _, exists := findTransactionByRef(stkReq.ExternalReference); exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("A transaction with reference %s already exists", stkReq.ExternalReference),
		})
		return
	}

	// Base64 encode credentials
	PAYHERO_USERNAME := viper.GetString("PAYHERO_USERNAME")
	PAYHERO_PASSWORD := viper.GetString("PAYHERO_PASSWORD")
	PAYHERO_CHANNEL_ID := viper.GetString("PAYHERO_CHANNEL_ID")
	CALLBACK_URL := viper.GetString("CALLBACK_URL")

	if PAYHERO_USERNAME == "" || PAYHERO_PASSWORD == "" || PAYHERO_CHANNEL_ID == "" || CALLBACK_URL == "" {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "PayHero credentials or configuration missing",
		})
		return
	}

	credentials := PAYHERO_USERNAME + ":" + PAYHERO_PASSWORD
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

	// Prepare request body to PayHero
	requestBody, err := json.Marshal(map[string]interface{}{
		"amount":             stkReq.Amount,
		"phone_number":       stkReq.PhoneNumber,
		"channel_id":         PAYHERO_CHANNEL_ID,
		"provider":           "m-pesa",
		"external_reference": stkReq.ExternalReference,
		"callback_url":       CALLBACK_URL,
	})
	if err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to marshal request body",
		})
		return
	}

	req, err := http.NewRequest("POST", "https://backend.payhero.co.ke/api/v2/payments", bytes.NewBuffer(requestBody))
	if err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to create request: " + err.Error(),
		})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedCredentials)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to make request: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	var payHeroResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payHeroResp); err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to decode response body",
		})
		return
	}

	// PayHero returns success boolean; respond consistently to frontend
	if successVal, ok := payHeroResp["success"].(bool); ok && successVal {
		// Return store.Response wrapping an STKResponse with the external_reference
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "STK push sent successfully",
			Data: STKResponse{
				ExternalReference: stkReq.ExternalReference,
				Data:              payHeroResp,
			},
		})
		return
	}

	// If we reach here it's not successful
	msg := "STK push was unsuccessful"
	if m, ok := payHeroResp["message"].(string); ok && m != "" {
		msg = msg + ": " + m
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: false,
		Message: msg,
	})
}

// --- StoreTransactionHandler: store successful transactions and update user if found ---
func StoreTransactionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		Response store.Transaction `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	transaction := reqBody.Response

	slog.Info("Received transaction", "transaction", transaction, "req body", reqBody, "raw", r.Body)
	if strings.TrimSpace(transaction.ExternalReference) == "" {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Missing external_reference",
		})
		return
	}

	// If transaction already stored, return early
	if _, exists := findTransactionByRef(transaction.ExternalReference); exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("Transaction with reference %s already exists", transaction.ExternalReference),
		})
		return
	}

	// Store only successful transactions
	if strings.EqualFold(transaction.Status, "Success") {
		if err := createTransaction(transaction); err != nil {
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to store transaction: " + err.Error(),
			})
			return
		}

		// Identify user
		parts := strings.Split(transaction.ExternalReference, "-")
		if len(parts) > 0 {
			identifier := parts[0]
			if _, userID, found := findUserByEmailOrUsername(identifier); found {
				store.Storage.Mu.Lock()
				u := store.Storage.Users[userID]

				// Update common fields
				u.Amount = transaction.Amount
				u.PhoneNumber = transaction.PhoneNumber
				u.UpdatedAt = time.Now()

				// Assign plan details based on amount
				now := time.Now()
				switch transaction.Amount {
				case 50:
					u.Plan = "student"
					u.PlanName = "Student Plan"
					u.Price = "50 Ksh"
					u.Duration = "5 hours"
					u.ExpireDuration = int64(5 * time.Hour.Seconds())
					u.ExpiryTimestamp = now.Add(5 * time.Hour).Unix()

				case 100:
					u.Plan = "hobbyist"
					u.PlanName = "Hobbyist Plan"
					u.Price = "100 Ksh"
					u.Duration = "24 hours"
					u.ExpireDuration = int64(24 * time.Hour.Seconds())
					u.ExpiryTimestamp = now.Add(24 * time.Hour).Unix()

				case 500:
					u.Plan = "pro"
					u.PlanName = "Pro Plan"
					u.Price = "500 Ksh"
					u.Duration = "1 week"
					u.ExpireDuration = int64((7 * 24) * time.Hour.Seconds())
					u.ExpiryTimestamp = now.Add(7 * 24 * time.Hour).Unix()
				}

				store.Storage.Users[userID] = u
				store.Storage.Mu.Unlock()
				_ = store.SaveStorage()
			}
		}
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Transaction processed successfully",
	})
}

func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store.Storage.Mu.RLock()
	transactions := make([]store.Transaction, 0, len(store.Storage.Transactions))
	for _, transaction := range store.Storage.Transactions {
		transactions = append(transactions, transaction)
	}
	store.Storage.Mu.RUnlock()

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Transactions retrieved successfully",
		Data: map[string]interface{}{
			"transactions": transactions,
			"count":        len(transactions),
		},
	})
}

func GetTransactionByRefHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	externalReference := params["external_reference"]

	transaction, exists := findTransactionByRef(externalReference)
	if !exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Transaction not found",
		})
		return
	}

	if transaction.Status != "Success" {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Transaction not found",
		})
		return
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Transaction retrieved successfully",
		Data:    transaction,
	})
}
