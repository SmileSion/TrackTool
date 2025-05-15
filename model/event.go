package model

// Event 事件结构体
type Event struct {
	ClientType  string `json:"clientType"`
	Site        string `json:"site"`
	EventType   string `json:"eventType"`
	EventDetail string `json:"eventDetail"`
	UserDetail  string `json:"userDetail"`
	Timestamp   int64  `json:"timestamp"`    // 后端生成的时间戳
	TimeStamp   int64  `json:"timeStamp"`   // 前端传入的时间戳（用于校验，仅存入Redis）
}

