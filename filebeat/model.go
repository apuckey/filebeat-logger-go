package filebeat

type StringMessage struct {
	Msg string `json:"msg"`
}

type LogEntry struct {
	Timestamp string      `json:"@timestamp"`
	Version   int64       `json:"@version"`
	Host      string      `json:"host"`
	Message   interface{} `json:"message"`
	Channel   string      `json:"channel"`
	Level     string      `json:"level"`
}
