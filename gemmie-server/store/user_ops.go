package store

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/imrany/gemmie/gemmie-server/internal/encrypt"
)

func CreateUser(user User) error {
	ctx := context.Background()

	if user.ResponseMode == "" {
		user.ResponseMode = ModesLightResponse
	}
	if user.UnsubscribeToken == "" {
		user.UnsubscribeToken = encrypt.GenerateUnsubscribeToken(user.ID)
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO users (
			id, username, email, password_hash, created_at, updated_at,
			preferences, work_function, theme, sync_enabled, plan, plan_name,
			amount, duration, phone_number, expiry_timestamp, expire_duration,
			price, response_mode, agree_to_terms, request_count_value,
			request_count_timestamp, email_verified, email_subscribed,
			verification_token, verification_token_expiry, unsubscribe_token, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
				  $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
	`

	_, err := DB.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt,
		user.UpdatedAt, user.Preferences, user.WorkFunction, user.Theme,
		user.SyncEnabled, user.Plan, user.PlanName, user.Amount, user.Duration,
		user.PhoneNumber, user.ExpiryTimestamp, user.ExpireDuration, user.Price,
		user.ResponseMode, user.AgreeToTerms, user.RequestCount.Count,
		user.RequestCount.Timestamp, user.EmailVerified, user.EmailSubscribed,
		user.VerificationToken, user.VerificationTokenExpiry, user.UnsubscribeToken,
		user.UserAgent,
	)

	return err
}

func GetUsers() ([]User, error) {
	ctx := context.Background()

	query := `
		SELECT id, username, email, password_hash, created_at, updated_at,
			   preferences, work_function, theme, sync_enabled, plan, plan_name,
			   amount, duration, phone_number, expiry_timestamp, expire_duration,
			   price, response_mode, agree_to_terms, request_count_value,
			   request_count_timestamp, email_verified, email_subscribed,
			   verification_token, verification_token_expiry, unsubscribe_token, user_agent
		FROM users
	`

	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var verificationTokenExpiry sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.CreatedAt, &user.UpdatedAt, &user.Preferences, &user.WorkFunction,
			&user.Theme, &user.SyncEnabled, &user.Plan, &user.PlanName, &user.Amount,
			&user.Duration, &user.PhoneNumber, &user.ExpiryTimestamp, &user.ExpireDuration,
			&user.Price, &user.ResponseMode, &user.AgreeToTerms, &user.RequestCount.Count,
			&user.RequestCount.Timestamp, &user.EmailVerified, &user.EmailSubscribed,
			&user.VerificationToken, &verificationTokenExpiry, &user.UnsubscribeToken,
			&user.UserAgent,
		)

		if verificationTokenExpiry.Valid {
			user.VerificationTokenExpiry = verificationTokenExpiry.Time
		}

		if err == sql.ErrNoRows {
			return nil, nil
		}

		if err != nil {
			return nil, err
		}

		// Get user transactions
		transactions, err := GetUserTransactions(user.PhoneNumber)
		if err != nil {
			slog.Error("Error getting user transactions", "user_id", user.ID, "error", err)
		}
		user.UserTransactions = transactions

		users = append(users, user)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserByID(userID string) (*User, error) {
	ctx := context.Background()

	query := `
		SELECT id, username, email, password_hash, created_at, updated_at,
			   preferences, work_function, theme, sync_enabled, plan, plan_name,
			   amount, duration, phone_number, expiry_timestamp, expire_duration,
			   price, response_mode, agree_to_terms, request_count_value,
			   request_count_timestamp, email_verified, email_subscribed,
			   verification_token, verification_token_expiry, unsubscribe_token, user_agent
		FROM users WHERE id = $1
	`

	user := &User{}
	var verificationTokenExpiry sql.NullTime

	err := DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.Preferences, &user.WorkFunction,
		&user.Theme, &user.SyncEnabled, &user.Plan, &user.PlanName, &user.Amount,
		&user.Duration, &user.PhoneNumber, &user.ExpiryTimestamp, &user.ExpireDuration,
		&user.Price, &user.ResponseMode, &user.AgreeToTerms, &user.RequestCount.Count,
		&user.RequestCount.Timestamp, &user.EmailVerified, &user.EmailSubscribed,
		&user.VerificationToken, &verificationTokenExpiry, &user.UnsubscribeToken,
		&user.UserAgent,
	)

	if verificationTokenExpiry.Valid {
		user.VerificationTokenExpiry = verificationTokenExpiry.Time
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	// Get user transactions
	transactions, err := GetUserTransactions(user.PhoneNumber)
	if err != nil {
		slog.Error("Error getting user transactions", "user_id", user.ID, "error", err)
	}
	user.UserTransactions = transactions

	return user, err
}

func GetUserByUsername(username string) (*User, error) {
	ctx := context.Background()

	query := `
		SELECT id, username, email, password_hash, created_at, updated_at,
			   preferences, work_function, theme, sync_enabled, plan, plan_name,
			   amount, duration, phone_number, expiry_timestamp, expire_duration,
			   price, response_mode, agree_to_terms, request_count_value,
			   request_count_timestamp, email_verified, email_subscribed,
			   verification_token, verification_token_expiry, unsubscribe_token,
			   user_agent
		FROM users WHERE username = $1
	`

	user := &User{}
	var verificationTokenExpiry sql.NullTime

	err := DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.Preferences, &user.WorkFunction,
		&user.Theme, &user.SyncEnabled, &user.Plan, &user.PlanName, &user.Amount,
		&user.Duration, &user.PhoneNumber, &user.ExpiryTimestamp, &user.ExpireDuration,
		&user.Price, &user.ResponseMode, &user.AgreeToTerms, &user.RequestCount.Count,
		&user.RequestCount.Timestamp, &user.EmailVerified, &user.EmailSubscribed,
		&user.VerificationToken, &verificationTokenExpiry, &user.UnsubscribeToken,
		&user.UserAgent,
	)

	if verificationTokenExpiry.Valid {
		user.VerificationTokenExpiry = verificationTokenExpiry.Time
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	// Get user transactions
	transactions, err := GetUserTransactions(user.PhoneNumber)
	if err != nil {
		slog.Error("Error getting user transactions", "user_id", user.ID, "error", err)
	}
	user.UserTransactions = transactions

	return user, err
}

func GetUserByEmail(email string) (*User, error) {
	ctx := context.Background()

	query := `
		SELECT id, username, email, password_hash, created_at, updated_at,
			   preferences, work_function, theme, sync_enabled, plan, plan_name,
			   amount, duration, phone_number, expiry_timestamp, expire_duration,
			   price, response_mode, agree_to_terms, request_count_value,
			   request_count_timestamp, email_verified, email_subscribed,
			   verification_token, verification_token_expiry, unsubscribe_token,
			   user_agent
		FROM users WHERE email = $1
	`

	user := &User{}
	var verificationTokenExpiry sql.NullTime

	err := DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.Preferences, &user.WorkFunction,
		&user.Theme, &user.SyncEnabled, &user.Plan, &user.PlanName, &user.Amount,
		&user.Duration, &user.PhoneNumber, &user.ExpiryTimestamp, &user.ExpireDuration,
		&user.Price, &user.ResponseMode, &user.AgreeToTerms, &user.RequestCount.Count,
		&user.RequestCount.Timestamp, &user.EmailVerified, &user.EmailSubscribed,
		&user.VerificationToken, &verificationTokenExpiry, &user.UnsubscribeToken,
		&user.UserAgent,
	)

	if verificationTokenExpiry.Valid {
		user.VerificationTokenExpiry = verificationTokenExpiry.Time
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	// Get user transactions
	transactions, err := GetUserTransactions(user.PhoneNumber)
	if err != nil {
		slog.Error("Error getting user transactions", "user_id", user.ID, "error", err)
	}
	user.UserTransactions = transactions

	return user, err
}

// GetUserTransactions retrieves all transactions for a user based on their phone number
func GetUserTransactions(phoneNumber string) ([]Transaction, error) {
	query := `
		SELECT
			t.id, t.external_reference, t.mpesa_receipt_number,
			t.checkout_request_id, t.merchant_request_id, t.amount,
			t.phone_number, t.result_code, t.result_description,
			t.status, t.created_at, t.updated_at
		FROM transactions t
		INNER JOIN users u ON t.phone_number = u.phone_number
		WHERE u.phone_number = $1
		ORDER BY t.created_at DESC
	`

	rows, err := DB.Query(query, phoneNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(
			&tx.ID, &tx.ExternalReference, &tx.MpesaReceiptNumber,
			&tx.CheckoutRequestID, &tx.MerchantRequestID, &tx.Amount,
			&tx.PhoneNumber, &tx.ResultCode, &tx.ResultDescription,
			&tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

// GetUserByPhoneWithTransactions retrieves a user and their transactions
func GetUserByPhoneWithTransactions(phoneNumber string) (*User, error) {
	// First get the user
	query := `
		SELECT
			id, username, email, password_hash, created_at, updated_at,
			preferences, work_function, theme, sync_enabled, plan, plan_name,
			amount, duration, phone_number, expiry_timestamp, expire_duration,
			price, response_mode, agree_to_terms, request_count_value,
			request_count_timestamp, email_verified, email_subscribed,
			verification_token, verification_token_expiry, unsubscribe_token,
			user_agent
		FROM users
		WHERE phone_number = $1
	`

	var user User
	var verificationTokenExpiry sql.NullTime

	err := DB.QueryRow(query, phoneNumber).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt, &user.Preferences, &user.WorkFunction,
		&user.Theme, &user.SyncEnabled, &user.Plan, &user.PlanName,
		&user.Amount, &user.Duration, &user.PhoneNumber, &user.ExpiryTimestamp,
		&user.ExpireDuration, &user.Price, &user.ResponseMode, &user.AgreeToTerms,
		&user.RequestCount.Count, &user.RequestCount.Timestamp,
		&user.EmailVerified, &user.EmailSubscribed, &user.VerificationToken,
		&verificationTokenExpiry, &user.UnsubscribeToken, &user.UserAgent,
	)
	if err != nil {
		return nil, err
	}

	if verificationTokenExpiry.Valid {
		user.VerificationTokenExpiry = verificationTokenExpiry.Time
	}

	// Get user transactions
	transactions, err := GetUserTransactions(phoneNumber)
	if err != nil {
		return nil, err
	}
	user.UserTransactions = transactions

	return &user, nil
}

// findUserByEmailOrUsername tries to find user by email or username in storage
func FindUserByEmailOrUsername(identifier string) (*User, string, bool) {
	userByEmail, err := GetUserByEmail(identifier)
	if err != nil {
		slog.Error("Error finding user by email", "email", identifier, "error", err)
	}

	if userByEmail != nil {
		// Get user transactions
		transactions, err := GetUserTransactions(userByEmail.PhoneNumber)
		if err != nil {
			slog.Error("Error getting user transactions", "user_id", userByEmail.ID, "error", err)
		}
		userByEmail.UserTransactions = transactions
		return userByEmail, userByEmail.ID, true
	}

	userByUsername, err := GetUserByUsername(identifier)
	if err != nil {
		slog.Error("Error finding user by username", "username", identifier, "error", err)
	}

	if userByUsername != nil {
		// Get user transactions
		transactions, err := GetUserTransactions(userByUsername.PhoneNumber)
		if err != nil {
			slog.Error("Error getting user transactions", "user_id", userByUsername.ID, "error", err)
		}
		userByUsername.UserTransactions = transactions
		return userByUsername, userByUsername.ID, true
	}

	// User not found
	return nil, "", false
}

func UpdateUser(user User) error {
	ctx := context.Background()
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users SET
			username = $2, email = $3, password_hash = $4, updated_at = $5,
			preferences = $6, work_function = $7, theme = $8, sync_enabled = $9,
			plan = $10, plan_name = $11, amount = $12, duration = $13,
			phone_number = $14, expiry_timestamp = $15, expire_duration = $16,
			price = $17, response_mode = $18, agree_to_terms = $19,
			request_count_value = $20, request_count_timestamp = $21,
			email_verified = $22, email_subscribed = $23, verification_token = $24,
			verification_token_expiry = $25, unsubscribe_token = $26, user_agent = $27
		WHERE id = $1
	`

	_, err := DB.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.UpdatedAt,
		user.Preferences, user.WorkFunction, user.Theme, user.SyncEnabled,
		user.Plan, user.PlanName, user.Amount, user.Duration, user.PhoneNumber,
		user.ExpiryTimestamp, user.ExpireDuration, user.Price, user.ResponseMode,
		user.AgreeToTerms, user.RequestCount.Count, user.RequestCount.Timestamp,
		user.EmailVerified, user.EmailSubscribed, user.VerificationToken,
		user.VerificationTokenExpiry, user.UnsubscribeToken, user.UserAgent,
	)

	return err
}

func DeleteUser(userID string) error {
	ctx := context.Background()
	_, err := DB.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	return err
}
