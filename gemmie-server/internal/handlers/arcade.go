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

// CreateArcadeHandler handles POST /api/arcades
func CreateArcadeHandler(w http.ResponseWriter, r *http.Request) {
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

	var req store.Arcade
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Create new arcade
	arcade := store.Arcade{
		ID:          encrypt.GenerateID(nil),
		UserId:      userID,
		Label:       req.Label,
		CreatedAt:   req.CreatedAt,
		UpdatedAt:   time.Now(),
		Code:        req.Code,
		Description: req.Description,
		CodeType:    req.CodeType,
		MessageId:   req.MessageId,
	}

	id, err := store.CreateArcade(&arcade)
	if err != nil {
		slog.Error("Failed to create arcade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to create arcade",
		})
		return
	}

	slog.Info("Arcade created successfully", "id", id, "user_id", userID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Arcade created successfully",
		Data:    id,
	})
}

// GetArcadeHandler handles GET /api/arcades/{id}
func GetArcadeHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	idStr := r.URL.Path[len("/api/arcades/"):]
	var id int64
	_, err = fmt.Sscan(idStr, &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid arcade ID",
		})
		return
	}

	arcade, err := store.GetArcadeById(id)
	if err != nil {
		slog.Error("Failed to get arcade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get arcade",
		})
		return
	}

	if arcade == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Arcade not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Arcade retrieved successfully",
		Data:    arcade,
	})
}

// GetArcadesHandler handles GET /api/arcades or GET /api/arcades/?option="html"
func GetArcadesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	optionStr := r.URL.Query().Get("option")
	if optionStr == "" {
		arcades, err := store.GetArcadesByOption(nil)
		if err != nil {
			slog.Error("Failed to get arcades", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to get arcades",
			})
			return
		}

		if len(arcades) == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Arcades not found",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Arcades retrieved successfully",
			Data:    arcades,
		})
		return
	}
	arcades, err := store.GetArcadesByOption(&optionStr)
	if err != nil {
		slog.Error("Failed to get arcades", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get arcades",
		})
		return
	}

	if len(arcades) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Arcades not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Arcades retrieved successfully",
		Data:    arcades,
	})
}

// UpdateArcadeHandler handles PUT /api/arcades/{id}
func UpdateArcadeHandler(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.URL.Path[len("/api/arcades/"):]
	var id int64
	_, err = fmt.Sscan(idStr, &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid arcade ID",
		})
		return
	}

	var req store.Arcade
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Get existing arcade
	arcade, err := store.GetArcadeById(id)
	if err != nil {
		slog.Error("Failed to get arcade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get arcade",
		})
		return
	}

	if arcade == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Arcade not found",
		})
		return
	}

	// Update arcade fields
	arcade.Label = req.Label
	if req.Code != "" {
		arcade.Code = req.Code
	}
	arcade.Description = req.Description
	if req.CodeType != "" {
		arcade.CodeType = req.CodeType
	}
	arcade.UpdatedAt = time.Now()

	updatedArcade, err := store.UpdateArcade(arcade)
	if err != nil {
		slog.Error("Failed to update arcade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to update arcade",
		})
		return
	}

	slog.Info("Arcade updated successfully", "id", arcade.ID, "user_id", userID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Arcade updated successfully",
		Data:    updatedArcade,
	})
}

// DeleteArcadeHandler handles DELETE /api/arcades/{id}
func DeleteArcadeHandler(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.URL.Path[len("/api/arcades/"):]
	var id int64
	_, err = fmt.Sscan(idStr, &id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Invalid arcade ID",
		})
		return
	}

	// Get existing arcade
	arcade, err := store.GetArcadeById(id)
	if err != nil {
		slog.Error("Failed to get arcade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to get arcade",
		})
		return
	}

	if arcade == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Arcade not found",
		})
		return
	}

	err = store.DeleteArcadeByID(id)
	if err != nil {
		slog.Error("Failed to delete arcade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Failed to delete arcade",
		})
		return
	}

	slog.Info("Arcade deleted successfully", "id", arcade.ID, "user_id", userID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: "Arcade deleted successfully",
	})
}
