package store

import (
	"context"
	"database/sql"
	"time"
)

func CreateUserData(userData UserData) error {
	ctx := context.Background()
	
	if userData.UpdatedAt.IsZero() {
		userData.UpdatedAt = time.Now()
	}
	query := `
		INSERT INTO user_data (user_id, chats, link_previews, current_chat_id, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			chats = EXCLUDED.chats,
			link_previews = EXCLUDED.link_previews,
			current_chat_id = EXCLUDED.current_chat_id,
			updated_at = EXCLUDED.updated_at
	`

	_, err := DB.ExecContext(ctx, query,
		userData.UserID, userData.Chats, userData.LinkPreviews,
		userData.CurrentChatID, userData.UpdatedAt,
	)

	return err
}

func GetUserData(userID string) (*UserData, error) {
	ctx := context.Background()
	
	query := `
		SELECT user_id, chats, link_previews, current_chat_id, updated_at
		FROM user_data WHERE user_id = $1
	`

	userData := &UserData{}
	err := DB.QueryRowContext(ctx, query, userID).Scan(
		&userData.UserID, &userData.Chats, &userData.LinkPreviews,
		&userData.CurrentChatID, &userData.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return userData, err
}

func UpdateUserData(userData UserData) error {
	userData.UpdatedAt = time.Now()
	return CreateUserData(userData) // Uses UPSERT
}