package generics

import (
	"errors"
	"net/http"

	// log "github.com/Sirupsen/logrus"
	// "github.com/facebookgo/stack"
	"github.com/gigawattio/go-commons/pkg/errorlib"
	"github.com/gigawattio/go-commons/pkg/web"
	"github.com/gigawattio/go-commons/pkg/web/helper"
)

var requestAlreadyHandledError = errors.New("already handled")

func RequestAlreadyHandled() error {
	return requestAlreadyHandledError
}

type (
	ObjectProcessorFunc  func() (object interface{}, err error)
	ObjectsProcessorFunc func(limit int64, offset int64) (object interface{}, n int, err error)
)

// GenericObjectEndpoint takes a function that produces a (result, error) tuple and runs it.
//
// statuses[0] may contain the success status code (optional, defaults to http.StatusOK).
// statuses[1] may contain the failure status code (optional, defaults to httpStatusInternalServerError).
//
func GenericObjectEndpoint(w http.ResponseWriter, req *http.Request, processorFunc ObjectProcessorFunc, statuses ...int) {
	var status int
	object, err := processorFunc()
	if err != nil {
		if err == requestAlreadyHandledError {
			return
		}
		if len(statuses) > 1 {
			status = statuses[1] // User-specified error status code.
		} else if err == errorlib.NotFoundError {
			status = http.StatusNotFound
		} else if err == errorlib.NotAuthorizedError {
			status = http.StatusForbidden
		} else {
			status = http.StatusInternalServerError
		}
		// stack.Caller(skip)
		// log.E
		logX.Error("X3 error running object processor on URI=%v status-code=%v: %s", req.RequestURI, status, err)
		web.RespondWithJson(w, status, web.JsonError(err))
		return
	}
	if len(statuses) > 0 {
		status = statuses[0] // User-specified success status code.
	} else {
		status = autoStatus(req)
	}
	web.RespondWithJson(w, status, object)
}

// GenericUserListEndpoint provides automatic pagination.
func GenericObjectsEndpoint(w http.ResponseWriter, req *http.Request, processorFunc ObjectsProcessorFunc, statuses ...int) {
	var (
		limit  = helper.Int64GetParam("limit", 10, req)
		offset = helper.Int64GetParam("offset", 0, req)
		status int
	)
	objects, n, err := processorFunc(limit, offset)
	if err != nil {
		if err == requestAlreadyHandledError {
			return
		}
		if len(statuses) > 1 {
			status = statuses[1] // User-specified error status code.
		} else if err == errorlib.NotFoundError {
			status = http.StatusNotFound
		} else if err == errorlib.NotAuthorizedError {
			status = http.StatusForbidden
		} else {
			status = http.StatusInternalServerError
		}
		logX.Error("error running listing processor for URI=%v limit=%v offset=%v: %s", req.RequestURI, limit, offset, err)
		web.RespondWithJson(w, status, web.JsonError(err))
		return
	}
	response := NewApiResponse(objects, n)
	if len(statuses) > 0 {
		status = statuses[0] // User-specified success status code.
	} else {
		status = autoStatus(req)
	}
	web.RespondWithJson(w, status, response)
}

func autoStatus(req *http.Request) (statusCode int) {
	switch req.Method {
	case "POST":
		statusCode = http.StatusCreated
	default:
		// i.e. HEAD, GET, PUT, PATCH, DELETE.
		statusCode = http.StatusOK
	}
	return
}
