package model

type ApiResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserID    string `json:"userID"`
	Wallet    int    `json:"wallet"`
	Asset     string `json:"asset"`
}

type Transaction struct {
	Type   TransactionType `json:"type"`
	Time   int64           `json:"time"`
	Amount int             `json:"amount"`
	UserID string          `json:"userID"`
}

type TransactionType string

const (
	DEBIT  TransactionType = "DEBIT"
	CREDIT TransactionType = "CREDIT"
)
