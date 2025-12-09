package store

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
)

// SaveSubscription saves or updates a subscription
// This already supports multiple devices per user via unique endpoint constraint
func SaveSubscription(ctx context.Context, userID string, sub SubscriptionRequest, userAgent string) error {
	query := `
        INSERT INTO push_subscriptions (user_id, endpoint, p256dh_key, auth_key, user_agent, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW())
        ON CONFLICT (endpoint)
        DO UPDATE SET
            user_id = EXCLUDED.user_id,
            p256dh_key = EXCLUDED.p256dh_key,
            auth_key = EXCLUDED.auth_key,
            user_agent = EXCLUDED.user_agent,
            updated_at = NOW()
    `
	_, err := DB.ExecContext(ctx, query, userID, sub.Endpoint, sub.Keys.P256dh, sub.Keys.Auth, userAgent)
	return err
}

// DeleteSubscription removes a subscription
func DeleteSubscription(ctx context.Context, endpoint string) error {
	query := `DELETE FROM push_subscriptions WHERE endpoint = $1`
	_, err := DB.ExecContext(ctx, query, endpoint)
	return err
}

// SubscriptionExists checks if a subscription exists for a given endpoint
func SubscriptionExists(ctx context.Context, endpoint string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM push_subscriptions WHERE endpoint = $1)`
	var exists bool
	err := DB.QueryRowContext(ctx, query, endpoint).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetSubscriptionByEndpoint retrieves a single subscription by endpoint
func GetSubscriptionByEndpoint(ctx context.Context, endpoint string) (*PushSubscription, error) {
	query := `
        SELECT user_id, endpoint, p256dh_key, auth_key, user_agent, created_at, updated_at
        FROM push_subscriptions
        WHERE endpoint = $1
    `
	var sub PushSubscription
	err := DB.QueryRowContext(ctx, query, endpoint).Scan(
		&sub.UserID, &sub.Endpoint, &sub.P256dhKey,
		&sub.AuthKey, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetSubscriptionsByUserID retrieves all subscriptions for a user
// This now properly returns ALL devices for a user
func GetSubscriptionsByUserID(ctx context.Context, userID string) ([]PushSubscription, error) {
	query := `
        SELECT user_id, endpoint, p256dh_key, auth_key, user_agent, created_at, updated_at
        FROM push_subscriptions
        WHERE user_id = $1
        ORDER BY created_at DESC
    `
	rows, err := DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []PushSubscription
	for rows.Next() {
		var sub PushSubscription
		err := rows.Scan(
			&sub.UserID, &sub.Endpoint, &sub.P256dhKey,
			&sub.AuthKey, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

// GetSubscriptionsByUserIDs retrieves subscriptions for multiple users
// This will return ALL devices for ALL specified users
func GetSubscriptionsByUserIDs(ctx context.Context, userIDs []string) ([]PushSubscription, error) {
	if len(userIDs) == 0 {
		return GetAllSubscriptions(ctx)
	}

	query := `
        SELECT user_id, endpoint, p256dh_key, auth_key, user_agent, created_at, updated_at
        FROM push_subscriptions
        WHERE user_id = ANY($1)
        ORDER BY created_at DESC
    `
	rows, err := DB.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []PushSubscription
	for rows.Next() {
		var sub PushSubscription
		err := rows.Scan(
			&sub.UserID, &sub.Endpoint, &sub.P256dhKey,
			&sub.AuthKey, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

// GetAllSubscriptions retrieves all subscriptions
func GetAllSubscriptions(ctx context.Context) ([]PushSubscription, error) {
	query := `
        SELECT user_id, endpoint, p256dh_key, auth_key, user_agent, created_at, updated_at
        FROM push_subscriptions
        ORDER BY created_at DESC
    `
	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []PushSubscription
	for rows.Next() {
		var sub PushSubscription
		err := rows.Scan(
			&sub.UserID, &sub.Endpoint, &sub.P256dhKey,
			&sub.AuthKey, &sub.UserAgent, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

// DeleteSubscriptionsByUserID removes all subscriptions for a user
// Useful when a user account is deleted
func DeleteSubscriptionsByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM push_subscriptions WHERE user_id = $1`
	_, err := DB.ExecContext(ctx, query, userID)
	return err
}
