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
	Type        TransactionType `json:"type"`
	Description string          `json:"description"`
	Time        int64           `json:"time"`
	Amount      int             `json:"amount"`
	UserID      string          `json:"userID"`
}

type TransactionType string

const (
	DEBIT  TransactionType = "DEBIT"
	CREDIT TransactionType = "CREDIT"
)

type GameSession struct {
	SessionID  string            `json:"sessionID"`
	UserId     string            `json:"userID"`
	GameStatus GameSessionStatus `json:"gameStatus"`
}

type RollSession struct {
	RollID        string            `json:"rollID"`
	GameSessionID string            `json:"gameSessionID"`
	UserID        string            `json:"userID"`
	WinningGame   int               `json:"winningGame"`
	FirstRoll     int               `json:"firstRow"`
	SecondRoll    int               `json:"secondRow"`
	RowStatus     GameSessionStatus `json:"rowStatus"`
}

type GameSessionStatus string

const (
	INPROGRESS GameSessionStatus = "IN_PROGRESS"
	COMPLETED  GameSessionStatus = "COMPLETED"
)
