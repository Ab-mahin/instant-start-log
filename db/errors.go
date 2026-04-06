package db

import "errors"

// ErrTagNotFound is returned when a tag does not exist for the given media.
var ErrTagNotFound = errors.New("tag not found")
