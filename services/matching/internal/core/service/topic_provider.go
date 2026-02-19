package service

import (
	"github.com/nepeta70/ride-hailing/internal/pkg/contracts"
	"github.com/nepeta70/ride-hailing/internal/pkg/ports"
)

type MatchingTopicProvider struct {
	eventTopicMap map[string]contracts.Topic
	allTopics     []string
}

func NewTopicProvider() ports.TopicProvider {
	return &MatchingTopicProvider{
		eventTopicMap: map[string]contracts.Topic{},
		allTopics:     []string{string(contracts.TopicMatching)},
	}
}

func (p *MatchingTopicProvider) AllTopics() []string {
	return p.allTopics
}

func (p *MatchingTopicProvider) GetTopicForEvent(eventType string) (contracts.Topic, error) {
	return contracts.TopicMatching, nil
}

var _ ports.TopicProvider = (*MatchingTopicProvider)(nil)
