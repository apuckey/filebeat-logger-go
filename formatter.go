package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Formatter interface {
	Format(ctx *MessageContext, message interface{}) string
}

type SimpleFormatter struct {
	FormatString string
}

func (f *SimpleFormatter) Format(ctx *MessageContext, message interface{}) string {
	switch msg := message.(type) {
	case string:
		return fmt.Sprintf(f.FormatString, ctx.Level, ctx.TimeStamp.Format(time.StampNano), ctx.File, ctx.Line, fmt.Sprintf(strings.TrimSuffix(msg, "\n")))
	default:
		js, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			return fmt.Sprintf(f.FormatString, ctx.Level, ctx.TimeStamp.Format(time.StampNano), ctx.File, ctx.Line, fmt.Sprintf(strings.TrimSuffix(string(js), "\n")))
		}
		return fmt.Sprintf(f.FormatString, ctx.Level, ctx.TimeStamp.Format(time.StampNano), ctx.File, ctx.Line, fmt.Sprintf(strings.TrimSuffix("Unable to format message", "\n")))
	}
}

var DefaultFormatter Formatter = &SimpleFormatter{
	FormatString: "[%[1]s %[2]s, %[3]s:%[4]d] %[5]s\n",
}
