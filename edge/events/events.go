package events

import (
	"fmt"
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
)

type EventFilterSet []eventLabelPair

type eventLabelPair struct {
	labels []string
	event  *pb.EventFilter
}

// New returns a slice of eventLabelPairs
func New(events *pb.EventFilters) *EventFilterSet {
	eSet := make(EventFilterSet, 0)
	for _, event := range events.GetEventFilters() {
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
func (eSet *EventFilterSet) Find(detectedLabels map[string]int32) *pb.EventFilter {
	for _, pair := range *eSet {
		if contains(detectedLabels, pair.labels) {
			return pair.event
		}
	}
	return nil
}

func (eSet *EventFilterSet) String() string {
	ret := "\n"
	for _, e := range *eSet {
		ret += fmt.Sprintf("e.labels: %v\ne.event: [%v],\n", e.labels, e.event)
	}
	return ret
}
