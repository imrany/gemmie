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
		INSERT INTO chats (id, user_id, title, created_at, updated_at, is_archived, last_message_at, is_private)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := DB.ExecContext(ctx, query,
		chat.ID, chat.UserId, chat.Title, chat.CreatedAt,
		chat.UpdatedAt, chat.IsArchived,
		chat.LastMessageAt, chat.IsPrivate,
	)

	return err
}

func GetChatById(ID string) (*Chat, error) {
	ctx := context.Background()

	query := `
		SELECT id, user_id, title, created_at, updated_at, is_archived, last_message_at, is_private
		FROM chats WHERE id = $1
	`

	chat := &Chat{}
	err := DB.QueryRowContext(ctx, query, ID).Scan(
		&chat.ID, &chat.UserId, &chat.Title, &chat.CreatedAt,
		&chat.UpdatedAt, &chat.IsArchived,
		&chat.LastMessageAt, &chat.IsPrivate,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// Fetch messages for the chat
	messages, err := GetMessagesByChatId(ID)
	if err != nil {
		return nil, err
	}
	chat.Messages = messages
	chat.MessageCount = len(chat.Messages)
	return chat, nil
}

func GetChatsByUserId(userId string) ([]Chat, error) {
	ctx := context.Background()

	// selects all chats belonging to a specific user, excluding chats where the user_id also exists in the arcades table.
	query := `
		SELECT id, user_id, title, created_at, updated_at,
		is_archived, last_message_at, is_private
		FROM chats c
		WHERE c.user_id = $1
		AND NOT EXISTS (SELECT 1 FROM arcades a WHERE a.user_id = c.user_id)
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
			&chat.UpdatedAt, &chat.IsArchived,
			&chat.LastMessageAt, &chat.IsPrivate,
		)
		if err != nil {
			return nil, err
		}

		// Fetch messages for the chat
		messages, err := GetMessagesByChatId(chat.ID)
		if err != nil {
			return nil, err
		}
		chat.MessageCount = len(messages)

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
		updated_at = $5, is_archived = $6, last_message_at = $7, is_private = $8
			WHERE id = $1
	`
	_, err := DB.ExecContext(ctx, query,
		&chat.ID, &chat.UserId, &chat.Title, &chat.CreatedAt,
		&chat.UpdatedAt, &chat.IsArchived,
		&chat.LastMessageAt, &chat.IsPrivate,
	)

	return err
}

func DeleteChatByID(ID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM messages WHERE chat_id = $1", ID)
	if err != nil {
		return err
	}
	_, err = DB.ExecContext(ctx, "DELETE FROM chats WHERE id = $1", ID)
	return err
}

func DeleteAllChatsByUserID(userID string) error {
	ctx := context.Background()

	// First, get all chat IDs for the user.
	query := "SELECT id FROM chats WHERE user_id = $1"
	rows, err := DB.QueryContext(ctx, query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var chatIDs []string
	for rows.Next() {
		var chatID string
		if err := rows.Scan(&chatID); err != nil {
			return err
		}
		chatIDs = append(chatIDs, chatID)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Delete messages for each chat.
	for _, chatID := range chatIDs {
		_, err = DB.ExecContext(ctx, "DELETE FROM messages WHERE chat_id = $1", chatID)
		if err != nil {
			return err
		}
	}

	// Finally, delete the chats themselves.
	_, err = DB.ExecContext(ctx, "DELETE FROM chats WHERE user_id = $1", userID)
	return err
}
