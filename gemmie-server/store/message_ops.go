package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

func CreateMessage(msg Message) error {
	ctx := context.Background()

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO messages (id, chat_id, prompt, response, created_at, model, references_ids)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := DB.ExecContext(ctx, query,
		msg.ID, msg.ChatId, msg.Prompt,
		msg.Response, msg.CreatedAt, msg.Model,
		pq.Array(msg.References),
	)

	return err
}

func GetMessagesByChatId(chatId string) ([]Message, error) {
	ctx := context.Background()

	query := `
	SELECT
		m.id, m.chat_id, m.prompt, m.response, m.created_at, m.model, m.references_ids
	FROM messages m
	WHERE m.chat_id = $1
	ORDER BY m.created_at ASC
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
			&msg.ID, &msg.ChatId, &msg.Prompt,
			&msg.Response, &msg.CreatedAt, &msg.Model,
			pq.Array(&msg.References),
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
		chat_id = $2, prompt = $3, response = $4, created_at = $5, model = $6, references_ids = $7
			WHERE id = $1
	`
	_, err := DB.ExecContext(ctx, query,
		msg.ID, msg.ChatId, msg.Prompt,
		msg.Response, msg.CreatedAt, msg.Model,
		pq.Array(msg.References),
	)

	return err
}

func GetMessageById(ID string) (*Message, error) {
	ctx := context.Background()

	query := `
	SELECT
		m.id, m.chat_id, m.prompt, m.response, m.created_at, m.model, m.references_ids
	FROM messages m WHERE m.id = $1
	`

	message := &Message{}
	err := DB.QueryRowContext(ctx, query, ID).Scan(
		&message.ID, &message.ChatId, &message.Prompt,
		&message.Response, &message.CreatedAt, &message.Model,
		pq.Array(&message.References),
	)

	if err == sql.ErrNoRows {
		return nil, err
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
