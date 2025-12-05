package store

import "context"

// SaveSubscription saves or updates a subscription
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

// GetSubscriptionsByUserID retrieves all subscriptions for a user
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
	rows, err := DB.QueryContext(ctx, query, userIDs)
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
