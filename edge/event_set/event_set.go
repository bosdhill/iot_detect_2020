package event_set

import (
	"fmt"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
)

type EventSet []eventLabelPair

type eventLabelPair struct {
	labels []string
	event  *pb.Event
}

// New returns a slice of eventLabelPairs
func New(events *pb.Events) *EventSet {
	eSet := make(EventSet, 0)
	for _, event := range events.GetEvents() {
		eSet = append(eSet, eventLabelPair{event.GetLabels(), event})
	}
	return &eSet
}

func contains(detectedLabels map[string]int32, labels []string) bool {
	for _, label := range labels {
		_, ok := detectedLabels[label]
		if !ok {
			return false
		}
	}
	return true
}

// Find returns the event if the labels in eventLabelPairs is a subset of detectedLabels
func (eSet *EventSet) Find(detectedLabels map[string]int32) *pb.Event {
	for _, pair := range *eSet {
		if contains(detectedLabels, pair.labels) {
			return pair.event
		}
	}
	return nil
}

func (eSet *EventSet) String() string {
	ret := "\n"
	for _, e := range *eSet {
		ret += fmt.Sprintf("e.labels: %v\ne.event: [%v],\n", e.labels, e.event)
	}
	return ret
}
