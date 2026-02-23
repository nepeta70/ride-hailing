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
func (r *RequestInfo) ToByteMap() map[string]string {
	return map[string]string{
		"sender-id":    r.Sender.ID.String(),
		"sender-role":  r.Sender.Role.String(),
		"sender-name":  r.Sender.Name,
		"request-id":   r.Trace.RequestID.String(),
		"timestamp":    strconv.FormatInt(r.Trace.Timestamp, 10),
		"retry-count":  strconv.Itoa(r.Trace.RetryCount),
		"country-code": r.Location.CountryCode,
		"app-version":  r.Client.AppVersion,
		"os":           r.Client.OS,
		"network":      r.Client.Network,
		"device-id":    r.Client.DeviceID,
	}
}

func NewInfoFromMap(headers map[string]string) (*RequestInfo, bool) {
	tmp := make(map[string]any)
	for k, v := range headers {
		tmp[k] = v
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
