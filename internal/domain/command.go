package domain

type SubscriptionAction string

const (
	SubscriptionActionSubscribe   SubscriptionAction = "SUBSCRIBE"
	SubscriptionActionUnsubscribe SubscriptionAction = "UNSUBSCRIBE"
)

type SubscriptionChannel string

const (
	SubscriptionChannelTrade SubscriptionChannel = "TRADE"
	SubscriptionChannelDepth SubscriptionChannel = "DEPTH"
)

type SubscriptionCommand struct {
	RequestID string
	Symbol    string
	Action    SubscriptionAction
	Channel   SubscriptionChannel
}
