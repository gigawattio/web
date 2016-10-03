package helper

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/nbio/hitch"
)

// ContextParam is a shortcut to get param values encoded in the url path.
// e.g. the "id" portion of /v1/apps/:id.
func ContextParam(name string, req *http.Request) string {
	param := hitch.Params(req).ByName(name)
	return param
}

// Int64ContextParam extracts an int64 value from the http context.
func Int64ContextParam(name string, req *http.Request) (int64, error) {
	paramString := hitch.Params(req).ByName(name)
	value, err := strconv.ParseInt(paramString, 10, 64)
	if err != nil {
		log.Info("Failed to parse paramString=%s into an int64: %s", paramString, err)
		return 0, fmt.Errorf("param lookup of '%v': %s", name, err)
	}
	return value, nil
}

func Int64GetParam(key string, defaultValue int64, req *http.Request) int64 {
	s := req.URL.Query().Get(key)
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}
