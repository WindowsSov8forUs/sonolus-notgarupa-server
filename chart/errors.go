package chart

import (
	"errors"
	"fmt"
)

var ErrUnexpectedBeat = errors.New("unexpected beat format")

const (
	MaxTimingGroups     = 64
	MaxScrollVelocities = 512
)

func tooManyTimingGroupsError(count int) error {
	return fmt.Errorf("too many timing groups: %d > %d", count, MaxTimingGroups)
}

func tooManyScrollVelocitiesError(count int) error {
	return fmt.Errorf("too many scroll velocities: %d > %d", count, MaxScrollVelocities)
}
