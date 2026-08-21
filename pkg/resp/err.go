package resp

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrProtocol           = errors.New("resp:")
	errInvalidHeader      = fmt.Errorf("%w: invalid header", ErrProtocol)
	errInvalidBody        = fmt.Errorf("%w: invalid body", ErrProtocol)
	errNegativeLength     = fmt.Errorf("%w: negative length", ErrProtocol)
	errUnexpectedSentinel = fmt.Errorf("%w: unexpected sentinel", ErrProtocol)
	errUnexpectedEOF      = fmt.Errorf("%w: %w", ErrProtocol, io.ErrUnexpectedEOF)
)
