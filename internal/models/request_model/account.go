package request_model

type CreateAccountRequest struct {
	Name     string `json:"name"     validate:"required"`
	Currency string `json:"currency" validate:"required,len=3"`
}

// Amounts are in the smallest currency unit (paise for INR, cents for USD).

type DepositRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

type TransferRequest struct {
	FromAccountID string `json:"from_account_id" validate:"required,uuid"`
	ToAccountID   string `json:"to_account_id"   validate:"required,uuid"`
	Amount        int64  `json:"amount"          validate:"required,gt=0"`
}

type ReverseTransferRequest struct {
	TransferID string `json:"transfer_id" validate:"required,uuid"`
}
