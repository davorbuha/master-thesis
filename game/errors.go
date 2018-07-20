package game

import (
	"errors"
	"fmt"
)

var (
	ErrInternal        = errors.New("Internal server error")
	ErrWrongAdminToken = errors.New("Wrong admin token")
)

type CommandNotAllowedError struct {
	Command string
}

func (e CommandNotAllowedError) Error() string {
	return fmt.Sprintf("Command not allowed %v", e.Command)
}
