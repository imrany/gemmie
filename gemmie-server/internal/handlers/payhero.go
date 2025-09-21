package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Response struct {
		ExternalReference string      `json:"external_reference"`
		Data              interface{} `json:"data"`
	} `json:"response"`
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

// Payment handlers
func SendSTKHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var stkReq STKRequest
	if err := json.NewDecoder(r.Body).Decode(&stkReq); err != nil {
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

	// Prepare request body
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

	// Create request
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

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to make request: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	var payHeroResp PayHeroResponse
	if err := json.NewDecoder(resp.Body).Decode(&payHeroResp); err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to decode response body",
		})
		return
	}

	if payHeroResp.Success {
		stkResponse := STKResponse{}
		stkResponse.Response.ExternalReference = stkReq.ExternalReference
		stkResponse.Response.Data = payHeroResp

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "STK push sent successfully",
			Data:    stkResponse.Response,
		})
	} else {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "STK push was unsuccessful",
		})
	}
}

func StoreTransactionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqBody struct {
		Response store.Transaction `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	transaction := reqBody.Response

	// Check if transaction already exists
	if _, exists := findTransactionByRef(transaction.ExternalReference); exists {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: fmt.Sprintf("A transaction with reference %s already exists", transaction.ExternalReference),
		})
		return
	}

	// Store transaction
	if err := createTransaction(transaction); err != nil {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to store transaction: " + err.Error(),
		})
		return
	}

	// Update order status if order exists
	if order, exists := FindOrderByRef(transaction.ExternalReference); exists {
		if err := UpdateOrderStatus(transaction.ExternalReference, "Paid"); err != nil {
			log.Printf("Failed to update order status: %v", err)
		}

		log.Printf("Order %s status updated to Paid", order.ID)
	}

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Transaction stored successfully",
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

	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Transaction retrieved successfully",
		Data:    transaction,
	})
}