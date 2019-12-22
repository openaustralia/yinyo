package commands

import "errors"

// ErrNotFound is the error for something not being found. Use this as a sentinal value
var ErrNotFound = errors.New("not found")
