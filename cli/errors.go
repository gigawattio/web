package cli

import (
	"errors"
)

var (
	AppNameRequiredError            = errors.New("AppName must not be empty")
	WebServiceProviderRequiredError = errors.New("WebServiceProvider must not be nil")
	NilWebServiceError              = errors.New("WebServiceProvider produced a nil WebService without any error")
)
