package event

type Event interface {
	isEvent()
}

type MessageEvent struct {
	Message string
}

func (MessageEvent) isEvent() {}

type ProgressPercentEvent struct {
	CurrentPercent float64
	TotalPercent   float64
}

func (ProgressPercentEvent) isEvent() {}

type ProgressCurrentEvent struct {
	Message string
}

func (ProgressCurrentEvent) isEvent() {}

type ProgressTotalEvent struct {
	Message string
}

func (ProgressTotalEvent) isEvent() {}

