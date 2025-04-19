package storage

import "errors"

var ErrNotFound = errors.New("resource not found")
var ErrDuplicateEntry = errors.New("duplicate entry")
