package store

import (
	"context"
	"fmt"
	"time"
)

// CreateArcade - creates an new arcade return id
func CreateArcade(arcade *Arcade) (*string, error) {
	ctx := context.Background()
	now := time.Now()
	query := `INSERT INTO arcades (user_id, code, label, code_type, description, created_at, updated_at, id, chat_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := DB.ExecContext(ctx, query, arcade.UserId, arcade.Code, arcade.Label, arcade.CodeType, arcade.Description, arcade.CreatedAt, now, arcade.ID, arcade.ChatId)
	if err != nil {
		return nil, err
	}
	return &arcade.ID, nil
}

// UpdateArcade - updates an arcade where user_id and id matches
func UpdateArcade(arcade *Arcade) (*Arcade, error) {
	ctx := context.Background()
	now := time.Now()
	if arcade.UpdatedAt.IsZero() {
		arcade.UpdatedAt = now
	}
	query := `UPDATE arcades SET code = $1, label = $2, code_type = $3, description = $4, updated_at = $5, chat_id = $6 WHERE user_id = $7 AND id = $8`
	result, err := DB.ExecContext(ctx, query, arcade.Code, arcade.Label, arcade.CodeType, arcade.Description, arcade.UpdatedAt, arcade.ChatId, arcade.UserId, arcade.ID)
	if err != nil {
		return nil, err
	}

	if count, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if count == 0 {
		return nil, fmt.Errorf("arcade not found")
	}

	return arcade, nil
}

// DeleteAllArcadesByUserID - Deletes all arcade by their user_id
func DeleteAllArcadesByUserID(userID string) error {
	ctx := context.Background()
	query := `
	BEGIN;
	DELETE FROM chats WHERE chat_id IN (SELECT chat_id FROM arcades WHERE user_id = $1);
	DELETE FROM arcades WHERE user_id = $1;
	COMMIT;
	`
	_, err := DB.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteArcadeByID - Deletes an arcade by its id
func DeleteArcadeByID(id int64) error {
	ctx := context.Background()
	query := `
	BEGIN;
	DELETE FROM chats WHERE chat_id IN (SELECT chat_id FROM arcades WHERE id = $1);
	DELETE FROM arcades WHERE id = $1;
	COMMIT;
	`
	_, err := DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// GetArcadeById - Gets an arcade by its id
func GetArcadeById(id int64) (*Arcade, error) {
	ctx := context.Background()
	query := `SELECT id, user_id, code, label, code_type, description, created_at, updated_at, chat_id FROM arcades WHERE id = $1`
	row := DB.QueryRowContext(ctx, query, id)
	var arcade Arcade
	err := row.Scan(&arcade.ID, &arcade.UserId, &arcade.Code, &arcade.Label, &arcade.CodeType, &arcade.Description, &arcade.CreatedAt, &arcade.UpdatedAt, &arcade.ChatId)
	if err != nil {
		return nil, err
	}
	return &arcade, nil
}

// GetArcadesByOption - gets all arcades that matches the option e.g user_id or code or code_type
func GetArcadesByOption(option any) ([]*Arcade, error) {
	ctx := context.Background()
	if option == nil {
		// gets all arcades
		query := `SELECT id, user_id, code, label, code_type, description, created_at, updated_at, chat_id FROM arcades ORDER BY updated_at DESC`
		ctx := context.Background()
		rows, err := DB.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var arcades []*Arcade
		for rows.Next() {
			var arcade Arcade
			err := rows.Scan(&arcade.ID, &arcade.UserId, &arcade.Code, &arcade.Label, &arcade.CodeType, &arcade.Description, &arcade.CreatedAt, &arcade.UpdatedAt, &arcade.ChatId)
			if err != nil {
				return nil, err
			}
			arcades = append(arcades, &arcade)
		}

		return arcades, nil
	}

	query := `SELECT id, user_id, code, label, code_type, description, created_at, updated_at, chat_id FROM arcades WHERE user_id = $1 OR code = $2 OR code_type = $3 ORDER BY updated_at DESC`
	rows, err := DB.QueryContext(ctx, query, option)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arcades []*Arcade
	for rows.Next() {
		var arcade Arcade
		err := rows.Scan(&arcade.ID, &arcade.UserId, &arcade.Code, &arcade.Label, &arcade.CodeType, &arcade.Description, &arcade.CreatedAt, &arcade.UpdatedAt, &arcade.ChatId)
		if err != nil {
			return nil, err
		}
		arcades = append(arcades, &arcade)
	}

	return arcades, nil
}
