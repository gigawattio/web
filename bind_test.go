package web

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Test_BindContentTypeParsing exercises `web.Bind()' for the ability to handle
// JSON, XML, and YAML content-types.
func Test_BindContentTypeParsing(t *testing.T) {
	type BindableMessage struct {
		Message string `json:"message"`
	}
	testHeaders := []string{
		"application/json",
		"Application/json",
		"application/Json",
		"APPLICATION/json",
		"APPLICATiON/json",
		"APPLICATION/Json",
		"APPLICATION/JSON",
		"application/JSON",
		"application/json ; charset=utf-8",
		"Application/json; charset=utf-8",
		"application/Json;CHARSET=utf-8",
		"APPLICATION/json;charset=UTF-8",
		"APPLICATiON/json",
		"  APPLICATION/Json",
		"APPLICATION/JSON	",
		"application/JSON      ",
		"application/x-yaml",
		"Application/x-Yaml; charset=utf-8",
		"Application/X-YAML; charset=garbagevalue",
		" application/xml ;;;",
		"text/xml;;;",
		"TeXt/Yaml; charset=utf-8",
		"TeXt/x-yaml;",
	}
	for i, headerValue := range testHeaders {
		for j, contentTypeKey := range []string{"content-type", "Content-Type", "Content-TYPe", "content-TYPE"} {
			msgContent := fmt.Sprintf("hello there i=%v, j=%v", i, j)
			var bodyContent string
			if strings.Contains(strings.ToLower(headerValue), "yaml") { // YAML test case.
				bodyContent = fmt.Sprintf(`---
message: %s
`, msgContent)
			} else if strings.Contains(strings.ToLower(headerValue), "xml") { // XML test case.
				bodyContent = fmt.Sprintf(`<BindableMessage><Message>%s</Message></BindableMessage>`, msgContent)
			} else { // Default to JSON test case.
				bodyContent = fmt.Sprintf(`{"message": "%s"}`, msgContent)
			}
			req, err := http.NewRequest("get", "/testing", bytes.NewBufferString(bodyContent))
			if err != nil {
				t.Fatalf("Error constructing HTTP request: %s", err)
			}
			req.Header.Add(contentTypeKey, headerValue)
			msg := &BindableMessage{}
			if err := Bind(req, msg); err != nil {
				t.Errorf(`Bind() failed for header "%s: %s": %s`, contentTypeKey, headerValue, err)
			}
			if msg.Message != msgContent {
				t.Errorf(`Bound message result value differs from raw input value; input="%s" msg.Message="%s"
raw body content:
	%s`, msgContent, msg.Message, bodyContent)
			}
		}
	}
}
