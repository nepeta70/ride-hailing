package domain

import "time"

type Location struct {
	EntityID   string    `json:"entity_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Geohash    string    `json:"geohash"`
	Accuracy   float32   `json:"accuracy"`
	Heading    float32   `json:"heading"`
	Speed      float32   `json:"speed"`
	CapturedAt time.Time `json:"captured_at"`
}

func (l *Location) NewLocation() *Location {
	return &Location{}
}
