package domain

type Payload struct {
	Command *SubscriptionCommand
	Trade   *TradeEvent
	Depth   *DepthEvent
}

func (payload Payload) Empty() bool {
	return payload.Command == nil && payload.Trade == nil && payload.Depth == nil
}
