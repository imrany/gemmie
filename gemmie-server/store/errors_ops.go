package store

import (
	"context"
	"database/sql"
	"time"
)

// PlatformError operations

func CreatePlatformError(error PlatformError) error {
	ctx := context.Background()

	if error.CreatedAt == "" {
		error.CreatedAt = time.Now().GoString()
	}

	query := `
		INSERT INTO platform_errors (
			id, user_id, message, description,
			action, status, context, severity,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := DB.ExecContext(ctx, query,
		error.ID, error.UserId, error.Message, error.Description,
		error.Action, error.Status, error.Context, error.Severity,
		error.CreatedAt, error.UpdatedAt,
	)

	return err
}

func GetPlatformErrors() ([]PlatformError, error) {
	ctx := context.Background()

	query := `
		SELECT id, user_id, message, description,
			action, status, context, severity,
			created_at, updated_at
		FROM platform_errors
		ORDER BY updated_at DESC
	`

	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errors []PlatformError

	for rows.Next() {
		var error PlatformError
		err := rows.Scan(
			&error.ID, &error.UserId, &error.Message, &error.Description,
			&error.Action, &error.Status, &error.Context, &error.Severity,
			&error.CreatedAt, &error.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		errors = append(errors, error)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return errors, nil
}

func GetPlatformErrorByID(errorID string) (*PlatformError, error) {
	ctx := context.Background()

	query := `
		SELECT id, user_id, message, description,
							action, status, context, severity,
							created_at, updated_at
		FROM platform_errors WHERE id = $1
	`

	error := &PlatformError{}
	err := DB.QueryRowContext(ctx, query, errorID).Scan(
		&error.ID, &error.UserId, &error.Message, &error.Description,
		&error.Action, &error.Status, &error.Context, &error.Severity,
		&error.CreatedAt, &error.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return error, err
}

func UpdatePlatformError(error PlatformError) error {
	ctx := context.Background()
	error.UpdatedAt = time.Now()

	query := `
		UPDATE platform_errors SET
		user_id = $2, message = $3, description = $4,
		action = $5, status = $6, context = $7, severity = $8,
		created_at = $9, updated_at = $10
		WHERE id = $1
	`

	_, err := DB.ExecContext(ctx, query,
		&error.ID, &error.UserId, &error.Message, &error.Description,
		&error.Action, &error.Status, &error.Context, &error.Severity,
		&error.CreatedAt, &error.UpdatedAt,
	)

	return err
}

func DeletePlatformErrorByID(errorID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM platform_errors WHERE id = $1", errorID)
	return err
}

func DeleteAllPlatformErrorByUserID(userID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM platform_errors WHERE user_id = $1", userID)
	return err
}
