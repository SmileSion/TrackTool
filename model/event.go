package model

// Event 事件结构体
type Event struct {
	ClientType  string `json:"client_type"`
	Site        string `json:"site"`
	EventType   string `json:"event_type"`
	EventDetail string `json:"event_detail"`
	UserDetail  string `json:"user_detail"`
	Timestamp   int64  `json:"timestamp"`    // 后端生成的时间戳
	TimeStamp   int64  `json:"time_stamp"`   // 前端传入的时间戳（用于校验，仅存入Redis）
}

