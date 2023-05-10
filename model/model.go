package model

type ApiResponse struct {
	Status  bool        `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserID    string `json:"user_i_d"`
	Wallet    int    `json:"wallet"`
	Asset     string `json:"asset"`
}
