package store

import (
	"context"
	"database/sql"
	"time"
)

func CreateMessage(msg Message) error {
	ctx := context.Background()

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	query := `
		INSERT INTO messages (id, chat_id, role, content, created_at, model)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := DB.ExecContext(ctx, query,
		msg.ID, msg.ChatId, msg.Role,
		msg.Content, msg.CreatedAt, msg.Model,
	)

	return err
}

func GetMessagesByChatId(chatId string) ([]Message, error) {
	ctx := context.Background()

	query := `
			SELECT id, chat_id, role, content, created_at, model
			FROM messages
			WHERE chat_id = $1
			ORDER BY created_at ASC
		`

	rows, err := DB.QueryContext(ctx, query, chatId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID, &msg.ChatId, &msg.Role,
			&msg.Content, &msg.CreatedAt, &msg.Model,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func UpdateMessage(msg Message) error {
	ctx := context.Background()

	query := `
	UPDATE messages SET
		chat_id = $2, role = $3, content = $4, created_at = $5, model = $6
 		WHERE id = $1
	`
	_, err := DB.ExecContext(ctx, query,
		&msg.ID, &msg.ChatId, &msg.Role,
		&msg.Content, &msg.CreatedAt, &msg.Model,
	)

	return err
}

func GetMessageById(ID string) (*Message, error) {
	ctx := context.Background()

	query := `
		SELECT id, chat_id, role, content, created_at, model
		FROM messages WHERE id = $1
	`

	message := &Message{}
	err := DB.QueryRowContext(ctx, query, ID).Scan(
		&message.ID, &message.ChatId, &message.Role,
		&message.Content, &message.CreatedAt, &message.Model,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return message, err
}

func DeleteMessageByID(ID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM messages WHERE id = $1", ID)
	return err
}

func DeleteAllMessageByChatID(chatID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM messages WHERE chat_id = $1", chatID)
	return err
}
