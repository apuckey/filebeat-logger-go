package filebeat

import (
	"encoding/json"
	"fmt"
	"github.com/apuckey/filebeat-logger-go"
	"net"
	"os"
	"time"
)

type Logger struct {
	conn      *net.UnixConn
	socket    string
	address   string
	host      string
	category  string
	channel   chan *LogEntry
	formatter logger.Formatter
}

func NewFilebeatLogger(socket, category string, formatter logger.Formatter, bufferSize int) (*Logger, error) {

	if socket == "" {
		fmt.Println(fmt.Sprintf("[FilebeatLogger]: socket cannot be empty"))
		os.Exit(1)
	}

	l := &Logger{
		socket:    socket,
		category:  category,
		channel:   make(chan *LogEntry, bufferSize),
		formatter: formatter,
	}

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(fmt.Sprintf("[FilebeatLogger]: unable to determine hostname: %s", err.Error()))
		os.Exit(1)
	}

	l.host = hostname

	l.connect()

	go l.sendLoop()

	return l, nil
}

func (s *Logger) connect() {
	var err error
	var conn *net.UnixConn

	conn, err = net.DialUnix("unix", nil, &net.UnixAddr{Name: s.socket, Net: "unix"})
	if err != nil {
		fmt.Println(fmt.Sprintf("[FilebeatLogger]: Unable to connect to logging socket: %s", err.Error()))
	}
	s.conn = conn
}

func (s *Logger) sendLoop() {

	defer func() {
		e := recover()
		if e != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[FilebeatLogger]: Restarting sender go routine.")
			go s.sendLoop()
		}
	}()

	for msg := range s.channel {
		if msg != nil {

			// json encode the message
			js, err := json.Marshal(msg)
			if err != nil {
				fmt.Println(fmt.Sprintf("[FilebeatLogger]: unable to marshal message to send to filebeat: %s", err.Error()))
				continue
			}

			// add a carriage return.
			js = append(js, "\n"...)

			length := len(js)

			for sent := 0; sent < length; {
				chunk, err := s.conn.Write(js[sent:])
				if err != nil {
					fmt.Println(fmt.Sprintf("[FilebeatLogger]: unable to send data to socket: %s", err.Error()))
					fmt.Println(fmt.Sprintf("%s", string(js)))
					break
				}
				sent += chunk
			}

		}
	}
}

func (s *Logger) sendOne(level string, message interface{}) {
	logEntry := &LogEntry{
		Timestamp: time.Now().UnixMilli(),
		Version:   1,
		Host:      s.host,
		Message:   message,
		Channel:   s.category,
		Level:     level,
	}
	s.channel <- logEntry
}

func (s *Logger) SetFormatter(f logger.Formatter) {
	s.formatter = f
}

func (s *Logger) Emit(ctx *logger.MessageContext, message interface{}) error {
	s.sendOne(ctx.Level, message)
	return nil
}
