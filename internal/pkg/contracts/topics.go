package contracts

type Topic string

const (
	TopicUser         Topic = "user"
	TopicLocation     Topic = "location"
	TopicRide         Topic = "ride"
	TopicDriver       Topic = "driver"
	TopicMatching     Topic = "matching"
	TopicNotification Topic = "notification"
	TopicRider        Topic = "rider"
)

func (t Topic) String() string {
	return string(t)
}
