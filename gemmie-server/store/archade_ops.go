package store

import (
	"context"
	"fmt"
	"time"
)

// CreateArchade - creates an new archade
func CreateArchade(archade *Archade) error {
	ctx := context.Background()
	now := time.Now()
	if archade.CreatedAt.IsZero() {
		archade.CreatedAt = now
	}
	query := `INSERT INTO archades (user_id, code, label, code_type, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.ExecContext(ctx, query, archade.UserId, archade.Code, archade.Label, archade.CodeType, archade.Description, archade.CreatedAt, now)
	if err != nil {
		return err
	}

	return nil
}

// UpdateArchade - updates an archade where user_id and id matches
func UpdateArchade(archade *Archade) (*Archade, error) {
	ctx := context.Background()
	now := time.Now()
	if archade.UpdatedAt.IsZero() {
		archade.UpdatedAt = now
	}
	query := `UPDATE archades SET code = ?, label = ?, code_type = ?, description = ?, updated_at = ? WHERE user_id = ? AND id = ?`
	result, err := DB.ExecContext(ctx, query, archade.Code, archade.Label, archade.CodeType, archade.Description, archade.UpdatedAt, archade.UserId)
	if err != nil {
		return nil, err
	}

	if count, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if count == 0 {
		return nil, fmt.Errorf("archade not found")
	}

	return archade, nil
}

// DeleteAllArchadesByUserID - Deletes all archade by their user_id
func DeleteAllArchadesByUserID(userID string) error {
	ctx := context.Background()
	query := `DELETE FROM archades WHERE user_id = ?`
	_, err := DB.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteArchadeByID - Deletes an archade by its id
func DeleteArchadeByID(id int64) error {
	ctx := context.Background()
	query := `DELETE FROM archades WHERE id = ?`
	_, err := DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// GetArchadeById - Gets an archade by its id
func GetArchadeById(id int64) (*Archade, error) {
	ctx := context.Background()
	query := `SELECT id, user_id, code, label, code_type, description, created_at, updated_at FROM archades WHERE id = ?`
	row := DB.QueryRowContext(ctx, query, id)
	var archade Archade
	err := row.Scan(&archade.ID, &archade.UserId, &archade.Code, &archade.Label, &archade.CodeType, &archade.Description, &archade.CreatedAt, &archade.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &archade, nil
}

// GetArchadesByOption - gets all archades that matches the option e.g user_id or code or code_type
func GetArchadesByOption(option any) ([]*Archade, error) {
	ctx := context.Background()
	if option == nil {
		// gets all archades
		query := `SELECT id, user_id, code, label, code_type, description, created_at, updated_at FROM archades`
		ctx := context.Background()
		rows, err := DB.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var archades []*Archade
		for rows.Next() {
			var archade Archade
			err := rows.Scan(&archade.ID, &archade.UserId, &archade.Code, &archade.Label, &archade.CodeType, &archade.Description, &archade.CreatedAt, &archade.UpdatedAt)
			if err != nil {
				return nil, err
			}
			archades = append(archades, &archade)
		}

		return archades, nil
	}

	query := `SELECT id, user_id, code, label, code_type, description, created_at, updated_at FROM archades WHERE user_id = ? OR code = ? OR code_type`
	rows, err := DB.QueryContext(ctx, query, option)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var archades []*Archade
	for rows.Next() {
		var archade Archade
		err := rows.Scan(&archade.ID, &archade.UserId, &archade.Code, &archade.Label, &archade.CodeType, &archade.Description, &archade.CreatedAt, &archade.UpdatedAt)
		if err != nil {
			return nil, err
		}
		archades = append(archades, &archade)
	}

	return archades, nil
}
