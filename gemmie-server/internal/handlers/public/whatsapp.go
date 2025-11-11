package public

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/imrany/gemmie/gemmie-server/store"
	"github.com/imrany/whats-email/pkg/whatsapp"
)

type WhatsappRequest struct {
	Message string   `json:"message"`
	To      []string `json:"to"`
}

// WhatsAppHandler handles WhatsApp requests - POST /api/whatsapp/send.
func WhatsAppHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")
	var request WhatsappRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if request.To == nil || request.Message == "" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(store.Response{
			Success: false,
			Message: "to or message field is empty, read more https://github.com/imrany/whats-email/pkg/whatsapp",
		})
		return
	}

	for _, val := range request.To {
		err := whatsapp.SendMessage(ctx, val, request.Message)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(store.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(store.Response{
		Success: true,
		Message: fmt.Sprintf("Message sent to: s%", request.To),
	})
}
