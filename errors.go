package shactions

import (
	"errors"
)

// ErrorCode is an internal, recoverable error caused by Smart Home Actions.
type ErrorCode struct {
	error
}

// Error returns the internal error code, or a blank string if nil.
func (ec ErrorCode) Error() string {
	if ec.error != nil {
		return ec.error.Error()
	}
	return ""
}

var (
	// ErrAuthExpired indicates credentials have expired.
	ErrAuthExpired = ErrorCode{
		error: errors.New("authExpired"),
	}

	// ErrAuthFailure is a general failure to authenticate.
	ErrAuthFailure = ErrorCode{
		error: errors.New("authFailure"),
	}

	// ErrDeviceOffline indicates the target is unreachable.
	ErrDeviceOffline = ErrorCode{
		error: errors.New("deviceOffline"),
	}

	// ErrTimeout indicates an internal timeout.
	ErrTimeout = ErrorCode{
		error: errors.New("timeout"),
	}

	// ErrDeviceTurnedOff indicates the device is turned hard off (different than unreachable).
	ErrDeviceTurnedOff = ErrorCode{
		error: errors.New("deviceTurnedOff"),
	}

	// ErrDeviceNotFound reports that the device doesn't exist on the partner's side. This indicates an internal error such as data synchronization or a race condition.
	ErrDeviceNotFound = ErrorCode{
		error: errors.New("deviceNotFound"),
	}

	// ErrValueOutOfRange indicates the range in parameters is out of bounds, such as the temperature of an AC unit.
	ErrValueOutOfRange = ErrorCode{
		error: errors.New("valueOutOfRange"),
	}

	// ErrNotSupported indicates the command or its parameters are unsupported (this should generally not happen, as traits and business logic should prevent it).
	ErrNotSupported = ErrorCode{
		error: errors.New("notSupported"),
	}

	// ErrProtocolError is a general failure in processing the request.
	ErrProtocolError = ErrorCode{
		error: errors.New("protocolError"),
	}

	// ErrUnknownError is for any other type of error, although in practice should be treated as a TODO.
	ErrUnknownError = ErrorCode{
		error: errors.New("unknownError"),
	}
)
