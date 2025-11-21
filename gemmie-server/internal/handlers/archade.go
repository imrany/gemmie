package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
	"github.com/imrany/gemmie/gemmie-server/store"
)

// CreateArchadeHandler handles POST /api/archades
func CreateArchadeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID header required",
		})
		return
	}

	// Verify user exists
	user, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Database error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	var req store.Archade
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Create new archade
	archade := store.Archade{
		ID:          encrypt.GenerateID(nil),
		UserId:      userID,
		Label:       req.Label,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Code:        req.Code,
		Description: req.Description,
		CodeType:    req.CodeType,
	}

	if err := store.CreateArchade(&archade); err != nil {
		slog.Error("Failed to create archade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to create archade",
		})
		return
	}

	slog.Info("Archade created successfully", "id", archade.ID, "user_id", userID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Archade created successfully",
		Data:    archade,
	})
}

// GetArchadeHandler handles GET /api/archades/{id}
func GetArchadeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID header required",
		})
		return
	}

	// Verify user exists
	user, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Database error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	idStr := r.URL.Path[len("/api/archades/"):]
	var id int64
	_, err = fmt.Sscan(idStr, &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid archade ID",
		})
		return
	}

	archade, err := store.GetArchadeById(id)
	if err != nil {
		slog.Error("Failed to get archade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get archade",
		})
		return
	}

	if archade == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Archade not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Archade retrieved successfully",
		Data:    archade,
	})
}

// GetArchadesHandler handles GET /api/archades or GET /api/archades/?option="html"
func GetArchadesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	optionStr := r.URL.Query().Get("option")
	if optionStr == "" {
		archades, err := store.GetArchadesByOption(nil)
		if err != nil {
			slog.Error("Failed to get archades", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to get archades",
			})
			return
		}

		if len(archades) == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Archades not found",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Archades retrieved successfully",
			Data:    archades,
		})
		return
	}
	archades, err := store.GetArchadesByOption(&optionStr)
	if err != nil {
		slog.Error("Failed to get archades", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get archades",
		})
		return
	}

	if len(archades) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Archades not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Archades retrieved successfully",
		Data:    archades,
	})
}

// UpdateArchadeHandler handles PUT /api/archades/{id}
func UpdateArchadeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID header required",
		})
		return
	}

	// Verify user exists
	user, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Database error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	idStr := r.URL.Path[len("/api/archades/"):]
	var id int64
	_, err = fmt.Sscan(idStr, &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid archade ID",
		})
		return
	}

	var req store.Archade
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Get existing archade
	archade, err := store.GetArchadeById(id)
	if err != nil {
		slog.Error("Failed to get archade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get archade",
		})
		return
	}

	if archade == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Archade not found",
		})
		return
	}

	// Update archade fields
	archade.Label = req.Label
	archade.Code = req.Code
	archade.Description = req.Description
	archade.CodeType = req.CodeType
	archade.UpdatedAt = time.Now()

	updatedArchade, err := store.UpdateArchade(archade)
	if err != nil {
		slog.Error("Failed to update archade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to update archade",
		})
		return
	}

	slog.Info("Archade updated successfully", "id", archade.ID, "user_id", userID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Archade updated successfully",
		Data:    updatedArchade,
	})
}

// DeleteArchadeHandler handles DELETE /api/archades/{id}
func DeleteArchadeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User ID header required",
		})
		return
	}

	// Verify user exists
	user, err := store.GetUserByID(userID)
	if err != nil {
		slog.Error("Database error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Database error",
		})
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "User not found",
		})
		return
	}

	idStr := r.URL.Path[len("/api/archades/"):]
	var id int64
	_, err = fmt.Sscan(idStr, &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid archade ID",
		})
		return
	}

	// Get existing archade
	archade, err := store.GetArchadeById(id)
	if err != nil {
		slog.Error("Failed to get archade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get archade",
		})
		return
	}

	if archade == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Archade not found",
		})
		return
	}

	err = store.DeleteArchadeByID(id)
	if err != nil {
		slog.Error("Failed to delete archade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to delete archade",
		})
		return
	}

	slog.Info("Archade deleted successfully", "id", archade.ID, "user_id", userID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Archade deleted successfully",
	})
}
