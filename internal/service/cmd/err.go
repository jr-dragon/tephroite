package cmd

import (
	"errors"
	"fmt"
)

var (
	ErrCommand    = errors.New("cmd:")
	errValidation = fmt.Errorf("%w: validation:", ErrCommand)
)

func errWrongNumberOfArguments(cmd Command) error {
	return fmt.Errorf("wrong number of arguments for '%s' command", cmd.String())
}
