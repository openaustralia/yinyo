package commands

import "errors"

// ErrNotFound is the error for something not being found. Use this as a sentinal value
var ErrNotFound = errors.New("not found")

// ErrAppNotAvailable is the error you get when you try to start a run but the code
// hasn't yet been uploaded
var ErrAppNotAvailable = errors.New("app not available")

// ErrArchiveFormat is the error you get trying to upload an archive with a bad format
var ErrArchiveFormat = errors.New("archive format")
