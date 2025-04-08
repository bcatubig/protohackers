package server

import "errors"

var ErrServerClosed = errors.New("tcp: server closed")
