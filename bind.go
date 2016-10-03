package web

import (
	"fmt"
	"net/http"
	"strings"
)

// Bind automatically deserializes the request body into the specified value.
//
// `value` must be a pointer to the value to be deserialized.
//
// The decoder is automatically selected based on the "Content-Type" header in
// the request.
//
// Currently supported content types are:
//     - JSON
//     - XML
//     - YAML
func Bind(req *http.Request, value interface{}) (err error) {
	full := req.Header.Get("Content-Type")
	// Support impure content-type values such as "application/json; charset=utf-8".
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(full, ";")[0]))
	switch {
	case contentType == MimeJson:
		err = DecodeJson(req.Body, value)
	case contentType == MimeXml || contentType == MimeXml2:
		err = DecodeXml(req.Body, value)
	case contentType == MimeYaml || contentType == MimeYaml2 || contentType == MimeYaml3:
		err = DecodeYaml(req.Body, value)
	default:
		err = fmt.Errorf("bind failed; unable to handle content-type=%s", contentType)
	}
	return
}
