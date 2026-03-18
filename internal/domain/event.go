package domain

type TradeEvent struct {
	RequestID    string
	Symbol       string
	EventTimeMS  int64
	TradeID      int64
	Price        string
	Quantity     string
	IsBuyerMaker bool
}

type DepthEvent struct {
	RequestID    string
	Symbol       string
	EventTimeMS  int64
	LastUpdateID int64
	Bids         []DepthLevel
	Asks         []DepthLevel
}

type DepthLevel struct {
	Price    string
	Quantity string
}
