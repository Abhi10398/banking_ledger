package service_model

import "github.com/google/uuid"

// TransferResult is returned by ExecuteTransfer so the caller gets the new
// balances without a second DB round-trip.
type TransferResult struct {
	TransferID     uuid.UUID `json:"transfer_id"`
	NewFromBalance int64     `json:"new_from_balance"`
	NewToBalance   int64     `json:"new_to_balance"`
}

// ReversalResult is returned by ReverseTransfer.
type ReversalResult struct {
	ReversalID         uuid.UUID `json:"reversal_id"`
	OriginalTransferID uuid.UUID `json:"original_transfer_id"`
}
