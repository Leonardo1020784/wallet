package models

type TransactionRequest struct {
    UserID string  `json:"user_id"`
    Amount float64 `json:"amount"`
}

type BalanceResponse struct {
    Email   string  `json:"email"`
    Balance float64 `json:"balance"`
}

type TransferRequest struct {
    FromUserID      string  `json:"from_user_id"`
    ToUserID        string  `json:"to_user_id"`
    Amount          float64 `json:"amount"`
}
