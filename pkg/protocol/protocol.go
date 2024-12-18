package protocol

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"
)

// RequestType represents valid command types.
type RequestType string

const (
	RequestGet     RequestType = "GET"
	RequestPut     RequestType = "PUT"
	RequestDel     RequestType = "DEL"
	RequestSub     RequestType = "SUB"
	RequestUnsub   RequestType = "UNSUB"
	RequestUnknown RequestType = "UNKNOWN"
)

type Request struct {
	Request   RequestType
	RequestID string
	ClientID  string
	Key       string
	Data      []byte // Decoded data (if present)
}

func parseRequestType(input string) RequestType {
	switch RequestType(input) {
	case RequestGet, RequestPut, RequestDel, RequestSub, RequestUnsub:
		return RequestType(input)
	default:
		return RequestUnknown
	}
}

func UnmarshalRequest(input string) (*Request, error) {
	parts := strings.Fields(input)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid input: expected at least 4 fields, got %d", len(parts))
	}

	cmdType := parseRequestType(parts[0])

	cmd := &Request{
		Request:   cmdType,
		RequestID: parts[1],
		ClientID:  parts[2],
		Key:       parts[3],
	}

	if len(parts) > 4 {
		data, err := base64.StdEncoding.DecodeString(parts[4])
		if err != nil {
			return nil, fmt.Errorf("invalid base64 data: %w", err)
		}
		cmd.Data = data
	}

	return cmd, nil
}

func MarshalRequest(cmd *Request) string {
	switch cmd.Request {
	case RequestGet, RequestDel, RequestSub, RequestUnsub:
		return fmt.Sprintf("%s %s %s %s", string(cmd.Request), cmd.RequestID, cmd.ClientID, cmd.Key)
	case RequestPut:
		data := base64.StdEncoding.EncodeToString(cmd.Data)
		return fmt.Sprintf("%s %s %s %s %s", string(cmd.Request), cmd.RequestID, cmd.ClientID, cmd.Key, data)
	default:
		return string(RequestUnknown)
	}

}

type Response struct {
	RequestID string
	Status    bool
	Data      []byte // Decoded data (if present)
}

func UnmarshalResponse(input string) (*Response, error) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid input: expected at least 2 fields, got %d", len(parts))
	}

	status := false
	switch parts[1] {
	case "OK":
		status = true
	case "FAILED":
		status = false
	default:
		return nil, fmt.Errorf("invalid input: status is not recognized. Expected OK or FAILED got %s", parts[1])
	}

	data := []byte{}
	var err error
	if len(parts) > 2 {
		encoded := parts[2]
		data, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("Error decoding Base64: %v", err)
		}
	}

	return &Response{
		RequestID: parts[0],
		Status:    status,
		Data:      data,
	}, nil
}

func MarshalResponse(resp *Response) string {
	status := ""
	if resp.Status {
		status = "OK"
	} else {
		status = "FAILED"
	}

	if len(resp.Data) > 0 {
		encoded := base64.StdEncoding.EncodeToString(resp.Data)
		return fmt.Sprintf("%s %s %s", resp.RequestID, status, encoded)
	} else {
		return fmt.Sprintf("%s %s", resp.RequestID, status)
	}
}

func MarshalResponseDebug(resp *Response) string {
	status := ""
	if resp.Status {
		status = "OK"
	} else {
		status = "FAILED"
	}

	if utf8.Valid(resp.Data) {
		return fmt.Sprintf("%s %s %s", resp.RequestID, status, string(resp.Data))
	} else {
		return fmt.Sprintf("%s %s %v", resp.RequestID, status, resp.Data)

	}
}

type NotifcationType string

const (
	NotificationHello    NotifcationType = "HELLO"
	NotificationUpdate   NotifcationType = "UPDATE"
	NotificationClose    NotifcationType = "CLOSE"
	NotificationNotFound NotifcationType = "NOTFOUND"
	NotificationUnkown   NotifcationType = "UNKNOWN"
)

type Notification struct {
	Type NotifcationType
	Key  string
	Data []byte
}

func MarshalNotification(n *Notification) string {
	if len(n.Data) > 0 {
		encoded := base64.StdEncoding.EncodeToString(n.Data)
		return fmt.Sprintf("%s %s %s", string(n.Type), n.Key, encoded)
	} else {
		return fmt.Sprintf("%s %s", string(n.Type), n.Key)
	}
}

func parseNotificationType(input string) NotifcationType {
	switch NotifcationType(input) {
	case NotificationHello, NotificationUpdate, NotificationClose, NotificationNotFound:
		return NotifcationType(input)
	default:
		return NotificationUnkown
	}
}

func UnmarshalNotification(input string) (*Notification, error) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid input: expected at least 2 fields, got %d", len(parts))
	}

	n := &Notification{
		Type: parseNotificationType(parts[0]),
		Key:  parts[1],
	}

	if len(parts) > 2 {
		data, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid base64 data: %w", err)
		}
		n.Data = data
	}

	return n, nil
}
