package cli

import (
	"errors"
)

var (
	AppNameRequiredError    = errors.New("AppName must not be empty")
	WebServiceRequiredError = errors.New("WebService must not be nil")
)
