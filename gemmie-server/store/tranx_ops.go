package store

import (
	"context"
	"database/sql"
	"time"
)

// Transaction operations

func CreateTransaction(tx Transaction) error {
	ctx := context.Background()
	
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}
	if tx.UpdatedAt.IsZero() {
		tx.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO transactions (
			id, external_reference, mpesa_receipt_number, checkout_request_id,
			merchant_request_id, amount, phone_number, result_code,
			result_description, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := DB.ExecContext(ctx, query,
		tx.ID, tx.ExternalReference, tx.MpesaReceiptNumber, tx.CheckoutRequestID,
		tx.MerchantRequestID, tx.Amount, tx.PhoneNumber, tx.ResultCode,
		tx.ResultDescription, tx.Status, tx.CreatedAt, tx.UpdatedAt,
	)

	return err
}

func GetTransactions() ([]Transaction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, external_reference, mpesa_receipt_number, checkout_request_id,
			   merchant_request_id, amount, phone_number, result_code,
			   result_description, status, created_at, updated_at
		FROM transactions
		ORDER BY created_at DESC
	`

	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction

	for rows.Next() {
		var tx Transaction
		err := rows.Scan(
			&tx.ID, &tx.ExternalReference, &tx.MpesaReceiptNumber, &tx.CheckoutRequestID,
			&tx.MerchantRequestID, &tx.Amount, &tx.PhoneNumber, &tx.ResultCode,
			&tx.ResultDescription, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func GetTransactionByID(txID string) (*Transaction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, external_reference, mpesa_receipt_number, checkout_request_id,
			   merchant_request_id, amount, phone_number, result_code,
			   result_description, status, created_at, updated_at
		FROM transactions WHERE id = $1
	`

	tx := &Transaction{}
	err := DB.QueryRowContext(ctx, query, txID).Scan(
		&tx.ID, &tx.ExternalReference, &tx.MpesaReceiptNumber, &tx.CheckoutRequestID,
		&tx.MerchantRequestID, &tx.Amount, &tx.PhoneNumber, &tx.ResultCode,
		&tx.ResultDescription, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return tx, err
}

func GetTransactionByExtRef(ref string) (*Transaction, error) {
	ctx := context.Background()
	
	query := `
		SELECT id, external_reference, mpesa_receipt_number, checkout_request_id,
			   merchant_request_id, amount, phone_number, result_code,
			   result_description, status, created_at, updated_at
		FROM transactions WHERE external_reference = $1
	`

	tx := &Transaction{}
	err := DB.QueryRowContext(ctx, query, ref).Scan(
		&tx.ID, &tx.ExternalReference, &tx.MpesaReceiptNumber, &tx.CheckoutRequestID,
		&tx.MerchantRequestID, &tx.Amount, &tx.PhoneNumber, &tx.ResultCode,
		&tx.ResultDescription, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return tx, err
}

func UpdateTransaction(tx Transaction) error {
	ctx := context.Background()
	tx.UpdatedAt = time.Now()

	query := `
		UPDATE transactions SET
			external_reference = $2, mpesa_receipt_number = $3,
			checkout_request_id = $4, merchant_request_id = $5, amount = $6,
			phone_number = $7, result_code = $8, result_description = $9,
			status = $10, updated_at = $11
		WHERE id = $1
	`

	_, err := DB.ExecContext(ctx, query,
		tx.ID, tx.ExternalReference, tx.MpesaReceiptNumber, tx.CheckoutRequestID,
		tx.MerchantRequestID, tx.Amount, tx.PhoneNumber, tx.ResultCode,
		tx.ResultDescription, tx.Status, tx.UpdatedAt,
	)

	return err
}