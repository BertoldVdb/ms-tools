package mshal

import "errors"

var (
	ErrorUnknownDevice   = errors.New("Unsupported device found")
	ErrorInvalidResponse = errors.New("Received invalid response")
	ErrorReadNotAllowed  = errors.New("Memory can't be read")
	ErrorWriteNotAllowed = errors.New("Memory can't be written")
	ErrorTimeout         = errors.New("The operation did not complete in time")
	ErrorPatchFailed     = errors.New("Could not patch code")
	ErrorMissingFunction = errors.New("This function is not supported in this mode")
	ErrorNoAck           = errors.New("No ACK received")
)
