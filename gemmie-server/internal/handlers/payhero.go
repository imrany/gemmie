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

// PlanConfig represents a subscription plan
type PlanConfig struct {
	Name           string
	Price          string
	Duration       string
	ExpireDuration time.Duration
}

var planConfigs = map[int]PlanConfig{
	50: {
		Name:           "Student Plan",
		Price:          "50 Ksh",
		Duration:       "5 hours",
		ExpireDuration: 5 * time.Hour,
	},
	100: {
		Name:           "Hobbyist Plan",
		Price:          "100 Ksh",
		Duration:       "24 hours",
		ExpireDuration: 24 * time.Hour,
	},
	500: {
		Name:           "Pro Plan",
		Price:          "500 Ksh",
		Duration:       "1 week",
		ExpireDuration: 7 * 24 * time.Hour,
	},
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

	return store.SaveStorage()
}

// findUserByEmailOrUsername tries to find user by email or username in storage
func findUserByEmailOrUsername(identifier string) (*store.User, string, bool) {
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

// validatePayHeroConfig checks if all required PayHero configuration is present
func validatePayHeroConfig() error {
	required := []string{"PAYHERO_USERNAME", "PAYHERO_PASSWORD", "PAYHERO_CHANNEL_ID", "CALLBACK_URL"}
	var missing []string

	for _, key := range required {
		if viper.GetString(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing configuration: %s", strings.Join(missing, ", "))
	}
	return nil
}

// updateUserPlan updates user plan based on transaction amount
func updateUserPlan(userID string, transaction store.Transaction) error {
	plan, exists := planConfigs[transaction.Amount]
	if !exists {
		slog.Warn("Unknown plan amount", "amount", transaction.Amount, "transaction", transaction.ExternalReference)
		return nil // Don't fail, just log warning
	}

	store.Storage.Mu.Lock()

	u, exists := store.Storage.Users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Log before update for debugging
	planKey := getPlanKey(transaction.Amount)
	slog.Info("Updating user plan", 
		"userID", userID,
		"amount", transaction.Amount,
		"planKey", planKey,
		"planName", plan.Name,
		"oldPlan", u.Plan,
	)

	// Update user plan
	now := time.Now()
	u.Plan = planKey
	u.PlanName = plan.Name
	u.Price = plan.Price
	u.Duration = plan.Duration
	u.Amount = transaction.Amount
	u.PhoneNumber = transaction.PhoneNumber
	u.ExpireDuration = int64(plan.ExpireDuration.Seconds())
	u.ExpiryTimestamp = now.Add(plan.ExpireDuration).Unix()
	u.UpdatedAt = now

	store.Storage.Users[userID] = u
	
	// Log after update for confirmation
	slog.Info("User plan updated", 
		"userID", userID,
		"newPlan", u.Plan,
		"expiryTimestamp", u.ExpiryTimestamp,
	)
	
	store.Storage.Mu.Unlock()
	return store.SaveStorage()
}

func getPlanKey(amount int) string {
	switch amount {
	case 50:
		return "student"
	case 100:
		return "hobbyist"
	case 500:
		return "pro"
	default:
		return "unknown"
	}
}

// checkAndUpdateUserFromTransaction checks if user needs plan update based on existing transaction
func checkAndUpdateUserFromTransaction(transaction store.Transaction) {
	// Only process successful transactions
	if !strings.EqualFold(transaction.Status, "Success") {
		return
	}

	// Extract user identifier from external reference
	parts := strings.Split(transaction.ExternalReference, "-")
	if len(parts) == 0 {
		return
	}

	identifier := parts[0]
	_, userID, found := findUserByEmailOrUsername(identifier)
	if !found {
		return
	}

	store.Storage.Mu.RLock()
	user := store.Storage.Users[userID]
	store.Storage.Mu.RUnlock()

	// Check if user needs plan update
	shouldUpdate := false
	currentTime := time.Now().Unix()
	transactionTime := transaction.CreatedAt.Unix()

	// Check if plan details are missing or incomplete
	if user.Plan == "" || user.PlanName == "" || user.ExpiryTimestamp == 0 {
		shouldUpdate = true
		slog.Info("User missing plan details", 
			"userID", userID, 
			"plan", user.Plan, 
			"planName", user.PlanName,
			"expiryTimestamp", user.ExpiryTimestamp,
		)
	}

	// Check if transaction is within valid duration and user should still have active plan
	plan, planExists := planConfigs[transaction.Amount]
	if planExists && shouldUpdate {
		expectedExpiryTime := transactionTime + int64(plan.ExpireDuration.Seconds())
		
		// If current time is still within the expected duration, update the user
		if currentTime <= expectedExpiryTime {
			slog.Info("Updating user plan from existing transaction within valid duration", 
				"userID", userID,
				"transactionTime", transactionTime,
				"expectedExpiryTime", expectedExpiryTime,
				"currentTime", currentTime,
				"reference", transaction.ExternalReference,
			)
			
			if err := updateUserPlan(userID, transaction); err != nil {
				slog.Error("Failed to update user plan from transaction check", 
					"error", err, 
					"userID", userID, 
					"reference", transaction.ExternalReference,
				)
			}
		} else {
			slog.Info("Transaction expired, not updating user plan", 
				"userID", userID,
				"transactionTime", transactionTime,
				"expectedExpiryTime", expectedExpiryTime,
				"currentTime", currentTime,
				"reference", transaction.ExternalReference,
			)
		}
	}
}

// SendSTKHandler initiates STK push payment
func SendSTKHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var stkReq STKRequest
	if err := json.NewDecoder(r.Body).Decode(&stkReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(stkReq.ExternalReference) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "External reference is required",
		})
		return
	}

	if stkReq.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Amount must be greater than 0",
		})
		return
	}

	if strings.TrimSpace(stkReq.PhoneNumber) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Phone number is required",
		})
		return
	}

	// Check if transaction already exists
	if _, exists := findTransactionByRef(stkReq.ExternalReference); exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("Transaction with reference %s already exists", stkReq.ExternalReference),
		})
		return
	}

	// Validate PayHero configuration
	if err := validatePayHeroConfig(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "PayHero configuration error",
		})
		slog.Error("PayHero config validation failed", "error", err)
		return
	}

	// Prepare credentials
	credentials := viper.GetString("PAYHERO_USERNAME") + ":" + viper.GetString("PAYHERO_PASSWORD")
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

	// Prepare request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"amount":             stkReq.Amount,
		"phone_number":       stkReq.PhoneNumber,
		"channel_id":         viper.GetString("PAYHERO_CHANNEL_ID"),
		"provider":           "m-pesa",
		"external_reference": stkReq.ExternalReference,
		"callback_url":       viper.GetString("CALLBACK_URL"),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to prepare request",
		})
		slog.Error("Failed to marshal PayHero request", "error", err)
		return
	}

	// Make request to PayHero
	req, err := http.NewRequest("POST", "https://backend.payhero.co.ke/api/v2/payments", bytes.NewBuffer(requestBody))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to create request",
		})
		slog.Error("Failed to create PayHero request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+encodedCredentials)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Payment service unavailable",
		})
		slog.Error("PayHero request failed", "error", err, "reference", stkReq.ExternalReference)
		return
	}
	defer resp.Body.Close()

	var payHeroResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payHeroResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to process payment response",
		})
		slog.Error("Failed to decode PayHero response", "error", err, "status", resp.StatusCode)
		return
	}

	// Check PayHero response
	if successVal, ok := payHeroResp["success"].(bool); ok && successVal {
		slog.Info("STK push initiated successfully", "reference", stkReq.ExternalReference, "amount", stkReq.Amount)
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

	// Handle unsuccessful response
	msg := "STK push failed"
	if m, ok := payHeroResp["message"].(string); ok && m != "" {
		msg = fmt.Sprintf("%s: %s", msg, m)
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(store.Response{
		Success: false,
		Message: msg,
	})
	slog.Warn("STK push failed", "reference", stkReq.ExternalReference, "response", payHeroResp)
}

// StoreTransactionHandler processes payment callbacks and updates user plans
func StoreTransactionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		Response store.Transaction `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		slog.Error("Failed to decode transaction callback", "error", err)
		return
	}

	transaction := reqBody.Response

	slog.Info("Processing transaction callback", 
		"reference", transaction.ExternalReference,
		"status", transaction.Status,
		"amount", transaction.Amount,
	)

	if strings.TrimSpace(transaction.ExternalReference) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "External reference is required",
		})
		return
	}

	// Check if transaction already exists
	if _, exists := findTransactionByRef(transaction.ExternalReference); exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("Transaction with reference %s already exists", transaction.ExternalReference),
		})
		slog.Warn("Duplicate transaction callback", "reference", transaction.ExternalReference)
		return
	}

	// Store the transaction regardless of status for audit trail
	if err := createTransaction(transaction); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to store transaction",
		})
		slog.Error("Failed to store transaction", "error", err, "reference", transaction.ExternalReference)
		return
	}

	// Only update user plan for successful transactions
	if strings.EqualFold(transaction.Status, "Success") {
		parts := strings.Split(transaction.ExternalReference, "-")
		if len(parts) > 0 {
			identifier := parts[0]
			if _, userID, found := findUserByEmailOrUsername(identifier); found {
				if err := updateUserPlan(userID, transaction); err != nil {
					slog.Error("Failed to update user plan", 
						"error", err, 
						"userID", userID, 
						"reference", transaction.ExternalReference,
					)
					// Don't fail the entire request, transaction is already stored
				} else {
					slog.Info("User plan updated successfully", 
						"userID", userID, 
						"plan", getPlanKey(transaction.Amount),
						"amount", transaction.Amount,
					)
				}
			} else {
				slog.Warn("User not found for transaction", 
					"identifier", identifier, 
					"reference", transaction.ExternalReference,
				)
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
		
		// Check and update user from existing transaction if needed
		// This runs in the background and doesn't affect the response
		go checkAndUpdateUserFromTransaction(transaction)
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

	if strings.TrimSpace(externalReference) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "External reference is required",
		})
		return
	}

	transaction, exists := findTransactionByRef(externalReference)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Transaction not found",
		})
		return
	}

	// Check and update user from this transaction if needed
	go checkAndUpdateUserFromTransaction(*transaction)

	// Only return successful transactions to the client
	if !strings.EqualFold(transaction.Status, "Success") {
		w.WriteHeader(http.StatusNotFound)
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