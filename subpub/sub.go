package subpub

type Event struct {
	Type    string
	Payload interface{}
}

type Subscriber interface {
	// Notify must be able to handle async calls
	Notify(Event)
}

type SubPub struct {
	subscribers []Subscriber
}

func New() *SubPub {
	return &SubPub{}
}

func (s *SubPub) Subscribe(sub Subscriber) {
	s.subscribers = append(s.subscribers, sub)
}

func (s *SubPub) Broadcast(event Event) {
	for _, sub := range s.subscribers {
		go func() {
			defer func() {
				if p := recover(); p != nil {
					// handle panic
					// log panic
				}
			}()
			sub.Notify(event)
		}()
	}
}
