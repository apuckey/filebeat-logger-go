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
	channel   string
	logchan   chan *LogEntry
	formatter logger.Formatter
}

func NewFilebeatLogger(socket, channel string, formatter logger.Formatter, bufferSize int) (*Logger, error) {

	if socket == "" {
		fmt.Println(fmt.Sprintf("[FilebeatLogger]: socket cannot be empty"))
		os.Exit(1)
	}

	l := &Logger{
		socket:    socket,
		channel:   channel,
		logchan:   make(chan *LogEntry, bufferSize),
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

	for msg := range s.logchan {
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
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   1,
		Host:      s.host,
		Message:   message,
		Channel:   s.channel,
		Level:     level,
	}
	s.logchan <- logEntry
}

func (s *Logger) SetFormatter(f logger.Formatter) {
	s.formatter = f
}

func (s *Logger) Emit(ctx *logger.MessageContext, message interface{}) error {
	// convert string message into struct
	switch msg := message.(type) {
	case string:
		s.sendOne(ctx.Level, &StringMessage{
			Msg: msg,
		})
	default:
		s.sendOne(ctx.Level, msg)
	}
	return nil
}
