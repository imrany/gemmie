package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/imrany/gemmie/gemmie-server/store"
)

type PlatformErrorRequest struct {
	Errors []store.PlatformError `json:"errors"`
}

// ErrorsHandler - gets/post/delete user platform errors
func ErrorsHandler(w http.ResponseWriter, r *http.Request) {
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

	var req PlatformErrorRequest
	switch r.Method {
	case "GET":
		platformErrors, err := store.GetPlatformErrors()
		if err != nil {
			slog.Error("Failed to get all platform errors data", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to retrieve data",
			})
			return
		}

		if platformErrors == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "platform errors data not found",
			})
			return
		}

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Data retrieved successfully",
			Data: map[string]any{
				"platform_errors": platformErrors,
			},
		})
	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Validate required fields
		if len(req.Errors) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Errors are required",
			})
			return
		}

		for _, platformError := range req.Errors {
			if err := store.CreatePlatformError(platformError); err != nil {
				slog.Error("Failed to record platform error", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(store.Response{
					Success: false,
					Message: "Failed to record platform error, try again",
				})
				return
			}

			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(store.Response{
				Success: true,
				Message: "Platform error recorded successfully, wait for further assistance from our team",
			})
		}
	case "DELETE":
		if err := store.DeleteAllPlatformErrorByUserID(userID); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to delete platform errors for user: " + userID,
			})
			return
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Platform errors deleted",
		})
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
	}
}

// ErrorHandler - gets/deletes/update specified platforme error by id
func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errorID := r.URL.Query().Get("id")
	if errorID == "" {
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "error id required",
		})
		return
	}

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

	switch r.Method {
	case "GET":
		platformError, err := store.GetPlatformErrorByID(errorID)
		if err != nil {
			slog.Error("Failed to get platform errors data", "errorID", errorID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to retrieve data",
			})
			return
		}

		if platformError == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "platform error data not found",
			})
			return
		}

		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Data retrieved successfully",
			Data:    platformError,
		})
	case "DELETE":
		if err := store.DeletePlatformErrorByID(errorID); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to delete platform error",
			})
			return
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Platform error deleted",
		})
		return
	case "PUT":
		var req store.PlatformError
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		// Validate required fields
		if req.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Errors are required",
			})
			return
		}

		if err := store.UpdatePlatformError(req); err != nil {
			slog.Error("Failed to record platform error", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: "Failed to record platform error, try again",
			})
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(store.Response{
			Success: true,
			Message: "Platform error updated successfully",
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "Method not allowed",
		})
	}
}
