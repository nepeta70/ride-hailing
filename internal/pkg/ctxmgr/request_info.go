package ctxmgr

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/nepeta70/ride-hailing/internal/pkg/core/enums"
)

type RequestInfo struct {
	Sender   Sender       `json:"sender"`
	Trace    TraceInfo    `json:"trace"`
	Location LocationInfo `json:"location"`
	Client   ClientInfo   `json:"client"`
}

// ToByteMap exports all fields to a flat map for Kafka headers or gRPC metadata.
func (r *RequestInfo) ToByteMap() map[string][]byte {
	return map[string][]byte{
		"sender-id":    []byte(r.Sender.ID.String()),
		"sender-role":  []byte(r.Sender.Role.String()),
		"sender-name":  []byte(r.Sender.Name),
		"request-id":   []byte(r.Trace.RequestID.String()),
		"timestamp":    []byte(strconv.FormatInt(r.Trace.Timestamp, 10)),
		"retry-count":  []byte(strconv.Itoa(r.Trace.RetryCount)),
		"country-code": []byte(r.Location.CountryCode),
		"app-version":  []byte(r.Client.AppVersion),
		"os":           []byte(r.Client.OS),
		"network":      []byte(r.Client.Network),
		"device-id":    []byte(r.Client.DeviceID),
	}
}

func NewRequestInfoFromByteMap(headers map[string][]byte) (*RequestInfo, bool) {
	tmp := make(map[string]any)
	for k, v := range headers {
		tmp[k] = string(v)
	}

	rInfo := &RequestInfo{}
	jsonBytes, _ := json.Marshal(tmp)
	json.Unmarshal(jsonBytes, rInfo)

	return rInfo, true
}

type Sender struct {
	ID   uuid.UUID        `json:"sender-id"`
	Role enums.SenderType `json:"sender-role"`
	Name string           `json:"sender-name"`
}

type TraceInfo struct {
	RequestID  uuid.UUID `json:"request-id"`
	Timestamp  int64     `json:"timestamp"`
	RetryCount int       `json:"retry-count"`
}

type LocationInfo struct {
	CountryCode string `json:"country-code"`
}

type ClientInfo struct {
	AppVersion string `json:"app-version"`
	OS         string `json:"os"`
	Network    string `json:"network"`
	DeviceID   string `json:"device-id"`
}
