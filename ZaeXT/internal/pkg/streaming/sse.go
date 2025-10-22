package streaming

import (
	"fmt"
	"io"
)

type SSEEvent struct {
	Event string
	Data  string
	Id    string
	Retry int
}

func (e *SSEEvent) FormatAsSSE() []byte {
	var result string
	if e.Event != "" {
		result += fmt.Sprintf("event: %s\n", e.Event)
	}
	if e.Data != "" {
		result += fmt.Sprintf("data: %s\n", e.Data)
	}
	if e.Id != "" {
		result += fmt.Sprintf("id: %s\n", e.Id)
	}
	if e.Retry > 0 {
		result += fmt.Sprintf("retry: %d\n", e.Retry)
	}
	result += "\n"
	return []byte(result)
}

func SendSSEEvent(writer io.Writer, event *SSEEvent) (int, error) {
	return writer.Write(event.FormatAsSSE())
}
