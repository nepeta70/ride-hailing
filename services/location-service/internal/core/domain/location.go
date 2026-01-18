package domain

import "time"

type UserType string

const (
	UserTypeDriver    UserType = "driver"
	UserTypePassenger UserType = "passenger"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type UserLocation struct {
	UserID      string      `json:"user_id"`
	UserType    UserType    `json:"user_type"`
	Coordinates Coordinates `json:"coordinates"`
	Accuracy    float32     `json:"accuracy"`
	Heading     float32     `json:"heading"`
	Speed       float32     `json:"speed"`
	CapturedAt  time.Time   `json:"captured_at"`
}

func (l *UserLocation) NewLocation() *UserLocation {
	return &UserLocation{}
}
