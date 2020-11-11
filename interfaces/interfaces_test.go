package interfaces

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvent_Flags_Enum(t *testing.T) {
	flags := Event_METADATA
	assert.Equal(t, flags&Event_METADATA, Event_METADATA, "flags should be only Event_METADATA")
	assert.Equal(t, flags&Event_ANNOTATED == 0, true, "flags should be only Event_METADATA")
	assert.Equal(t, flags&Event_BOXES == 0, true, "flags should be only Event_METADATA")

	flags = Event_ANNOTATED | Event_BOXES
	assert.Equal(t, flags&Event_ANNOTATED != 0, true, "flags should contain Event_ANNOTATED")
	assert.Equal(t, flags&Event_BOXES != 0, true, "flags should contain Event_BOXES")
	assert.Equal(t, flags&Event_CONFIDENCE == 0, true, "flags should not contain Event_CONFIDENCE")

	flags = Event_CONFIDENCE | Event_ANNOTATED | Event_BOXES | Event_PERSIST | Event_IMAGE
	assert.Equal(t, flags&Event_CONFIDENCE != 0, true, "flags should contain Event_CONFIDENCE")
	assert.Equal(t, flags&Event_ANNOTATED != 0, true, "flags should contain Event_ANNOTATED")
	assert.Equal(t, flags&Event_BOXES != 0, true, "flags should contain Event_BOXES")
	assert.Equal(t, flags&Event_PERSIST != 0, true, "flags should contain Event_PERSIST")
	assert.Equal(t, flags&Event_IMAGE != 0, true, "flags should contain Event_IMAGE")
}
