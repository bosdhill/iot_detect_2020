package settrie

import (
	pb "github.com/bosdhill/iot_detect_2020/interfaces"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Key []string

func TestFind_String_Subset(t *testing.T) {
	tree := New()
	l := []string{"person","car"}
	e := pb.Event{
		Uid:             0,
		Labels:          l,
		LabelEvents:     nil,
		DistanceMeasure: 0,
		Flags:           0,
	}

	tree.Add(l, e)

	event, _ := tree.Find([]string{"person", "car"})

	assert.NotEqual(t, event, nil)

	event, _ = tree.Find([]string{"car", "person"})

	assert.NotEqual(t, event, nil)

	event, _ = tree.Find([]string{"car", "person", "bus"})

	assert.NotEqual(t, event, nil)

}