package filebeat

type LogEntry struct {
	Timestamp int64       `json:"@timestamp"`
	Version   int64       `json:"@version"`
	Host      string      `json:"host"`
	Message   interface{} `json:"message"`
	Type      string      `json:"type"`
	Channel   string      `json:"channel"`
	Level     string      `json:"level"`
}
