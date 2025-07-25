package api

import (
	"errors"
)

var (
	ErrNoKey       = errors.New("no such key in cache")
	ErrNoRow       = errors.New("no such row in db")
	ErrNoRedirects = errors.New("no redirects for such link")
)
