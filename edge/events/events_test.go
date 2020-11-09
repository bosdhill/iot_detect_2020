package events

import (
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"testing"
)

// TestFind_String_Subset (an example of a Table-Driven test) checks whether the event can be found in the event set
func TestFind_String_Subset(t *testing.T) {
	cases := []struct {
		description    string
		detectedLabels map[string]int
		event          *pb.Event
		expected       bool
	}{
		{
			description:    "The event labels are a subset of the detectedLabels",
			detectedLabels: map[string]int32{"bus": 1, "car": 2, "person": 1, "truck": 1},
			event: &pb.Event{
				LabelEvents: map[string]*pb.EventConditions{
					"person": {
						ConfThreshold: 0.30,
						Quantity:      1,
						QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
						Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
					},
					"bus": {
						ConfThreshold: 0.30,
						Quantity:      1,
						QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
						Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
					},
				},
				Labels:          []string{"person", "bus"},
				DistanceMeasure: pb.Event_DISTANCE_MEASURE_UNSPECIFIED,
				Flags:           uint32(pb.Event_METADATA),
			},
			expected: true,
		},
		{
			description:    "The event labels are a not subset of the detectedLabels",
			detectedLabels: map[string]int32{"car": 2, "person": 1, "truck": 1},
			event: &pb.Event{
				LabelEvents: map[string]*pb.EventConditions{
					"person": {
						ConfThreshold: 0.30,
						Quantity:      1,
						QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
						Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
					},
					"bus": {
						ConfThreshold: 0.30,
						Quantity:      1,
						QuantityBound: uint32(pb.EventConditions_GREATER | pb.EventConditions_EQUAL),
						Proximity:     pb.EventConditions_PROXIMITY_UNSPECIFIED,
					},
				},
				Labels:          []string{"person", "bus"},
				DistanceMeasure: pb.Event_DISTANCE_MEASURE_UNSPECIFIED,
				Flags:           uint32(pb.Event_METADATA),
			},
			expected: false,
		},
	}

	for _, c := range cases {
		events := &pb.Events{}
		events.Events = append(events.Events, c.event)

		eSet := New(events)
		e := eSet.Find(c.detectedLabels)

		got := e == c.event
		if c.expected != got {
			t.Errorf("%v", c.description)
		}
	}
}
