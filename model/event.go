package model

// Event 事件结构体
type Event struct {
	ClientType  string `json:"client_type"`
	Site        string `json:"site"`
	EventType   string `json:"event_type"`
	EventDetail string `json:"event_detail"`
	UserDetail  string `json:"user_detail"` 
	Timestamp   int64  `json:"timestamp"`
}
