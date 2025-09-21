package handlers

import (
	"fmt"
	"time"

	"github.com/imrany/gemmie/gemmie-server/store"
)

func FindOrderByRef(ref string) (*store.Order, bool) {
	store.Storage.Mu.RLock()
	defer store.Storage.Mu.RUnlock()

	for _, order := range store.Storage.Orders {
		if order.ExternalReference == ref {
			return &order, true
		}
	}
	return nil, false
}

func UpdateOrderStatus(ref, status string) error {
	store.Storage.Mu.Lock()
	defer store.Storage.Mu.Unlock()

	for id, order := range store.Storage.Orders {
		if order.ExternalReference == ref {
			order.Status = status
			order.UpdatedAt = time.Now()
			store.Storage.Orders[id] = order
			store.SaveStorage()
			return nil
		}
	}
	return fmt.Errorf("order not found with reference: %s", ref)
}