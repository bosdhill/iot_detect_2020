package event_set

import (
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestFind_String_Subset(t *testing.T) {
	event := pb.Event{
		Uid:             0,
		Labels:          nil,
		LabelEvents:     nil,
		DistanceMeasure: 0,
		Flags:           0,
	}

	assert.NotEqual(t, event, nil)
}