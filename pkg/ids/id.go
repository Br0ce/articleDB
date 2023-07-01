package ids

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("invalid id")
)

func UniqueID() string {
	return uuid.NewString()
}

func ValidID(id string) bool {
	if id == "" {
		return false
	}

	if _, err := uuid.Parse(id); err != nil {
		return false
	}

	return true
}
