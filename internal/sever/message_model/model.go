package model

type WebsocketReqMessage struct {
	Type      string `json:"type"`
	SessionID int    `json:"session_id"`
	Content   string `json:"content"`
	RoleName  string `json:"role_name"`
}

type WebsocketRespMessage struct {
	Type     string `json:"type"`
	RoleName string `json:"role_name"`
	Content  string `json:"content"`
}
