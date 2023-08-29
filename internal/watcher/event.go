package watcher

const (
	EventCreated  EventType = "created"
	EventModified EventType = "modified"
	EventRemoved  EventType = "removed"
)

type EventType string
type EventsChannel <-chan Event

type Event struct {
	FileName string
	Type     EventType
	IsDir    bool
}
