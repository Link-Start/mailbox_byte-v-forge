package eventbus

import "errors"

var (
	ErrEmptyEnvelope          = errors.New("event envelope is required")
	ErrEmptyPayloadType       = errors.New("event payload type is required")
	ErrMismatchedSubject      = errors.New("event subject does not match expected event")
	ErrMismatchedEventType    = errors.New("event metadata type does not match expected event")
	ErrMismatchedEventVersion = errors.New("event metadata version does not match expected event")
	ErrMismatchedPayloadType  = errors.New("event payload type does not match expected event")
)
