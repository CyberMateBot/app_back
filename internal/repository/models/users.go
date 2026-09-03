package models

import "errors"

var (
	ErrUserIsNotFound = errors.New("user not found")
)

type User struct {
	ID      int64
	Name    string
	Surname string
}
