package domain

import "github.com/nepeta70/ride-hailing/internal/pkg/errors"

var (
	// ErrLocationNotFound indicates that the requested location does not exist
	ErrLocationNotFound = errors.NewBusinessError("location not found")
)
