package store

import (
	"context"
	"database/sql"
	"time"
)

func CreateChat(chat Chat) error {
	ctx := context.Background()

	if chat.LastMessageAt.IsZero() {
		chat.LastMessageAt = time.Now()
	}
	query := `
		INSERT INTO chats (id, user_id, title, created_at, updated_at, is_archived, message_count, last_message_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := DB.ExecContext(ctx, query,
		chat.ID, chat.UserId, chat.Title, chat.CreatedAt,
		chat.UpdatedAt, chat.IsArchived, chat.MessageCount,
		chat.LastMessageAt,
	)

	return err
}

func GetChatById(ID string) (*Chat, error) {
	ctx := context.Background()

	query := `
		SELECT id, user_id, title, created_at, updated_at, is_archived, message_count, last_message_at
		FROM chats WHERE id = $1
	`

	chat := &Chat{}
	err := DB.QueryRowContext(ctx, query, ID).Scan(
		&chat.ID, &chat.UserId, &chat.Title, &chat.CreatedAt,
		&chat.UpdatedAt, &chat.IsArchived, &chat.MessageCount,
		&chat.LastMessageAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return chat, err
}

func GetChatsByUserId(userId string) ([]Chat, error) {
	ctx := context.Background()

	query := `
			SELECT id, user_id, title, created_at, updated_at,
			is_archived, message_count, last_message_at
			FROM chats
			WHERE user_id = $1
			ORDER BY updated_at DESC
		`

	rows, err := DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []Chat

	for rows.Next() {
		var chat Chat
		err := rows.Scan(
			&chat.ID, &chat.UserId, &chat.Title, &chat.CreatedAt,
			&chat.UpdatedAt, &chat.IsArchived, &chat.MessageCount,
			&chat.LastMessageAt,
		)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func UpdateChat(chat Chat) error {
	ctx := context.Background()
	chat.UpdatedAt = time.Now()

	query := `
	UPDATE chats SET
		user_id = $2, title = $3, created_at = $4,
		updated_at = $5, is_archived = $6, message_count = $7, last_message_at = $8
 		WHERE id = $1
	`
	_, err := DB.ExecContext(ctx, query,
		&chat.ID, &chat.UserId, &chat.Title, &chat.CreatedAt,
		&chat.UpdatedAt, &chat.IsArchived, &chat.MessageCount,
		&chat.LastMessageAt,
	)

	return err
}

func DeleteChatByID(ID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM chats WHERE id = $1", ID)
	return err
}

func DeleteAllChatsByUserID(userID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM chats WHERE user_id = $1", userID)
	return err
}
