package service

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type RideTopicProvider struct {
	eventTopicMap map[string]contracts.Topic
	allTopics     []string
}

func NewTopicProvider() ports.TopicProvider {
	return &RideTopicProvider{
		eventTopicMap: map[string]contracts.Topic{},
		allTopics:     []string{string(contracts.TopicRide)},
	}
}

func (p *RideTopicProvider) AllTopics() []string {
	return p.allTopics
}

func (p *RideTopicProvider) GetTopicForEvent(eventType string) (contracts.Topic, error) {
	return contracts.TopicRide, nil
}

var _ ports.TopicProvider = (*RideTopicProvider)(nil)
