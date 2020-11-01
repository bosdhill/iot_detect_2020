package event_set

import pb "github.com/bosdhill/iot_detect_2020/interfaces"

type EventSet []eventLabelPair

type eventLabelPair struct {
	labels []string
	event *pb.Event
}

// New returns a slice of eventLabelPairs
func New(events *pb.Events) *EventSet {
	eSet := make(EventSet, len(events.GetEvents()))
	for _, event := range events.GetEvents() {
		eSet = append(eSet, eventLabelPair{event.GetLabels(), event})
	}
	return &eSet
}

func contains(detectedLabels map[string]int , labels []string) bool {
	for _, label := range labels {
		_, ok := detectedLabels[label]
		if !ok {
			return false
		}
	}
	return true
}

// Find returns the event if the labels in eventLabelPairs is a subset of detectedLabels
func (eSet *EventSet) Find(detectedLabels map[string]int) *pb.Event {
	for _, pair := range *eSet {
		if contains(detectedLabels, pair.labels) {
			return pair.event
		}
	}
	return nil
}