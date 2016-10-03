package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var (
	listenAddr = "127.0.0.1:45678"

	indexData = []byte(`<!DOCTYPE html>
<html>
	<head>
		<title>My Homepage</title>
	</head>
	<body>
		Welcome!
	</body>
</html>`)

	jsonData    = []byte(`{"some":"data"}`)
	dataStrings = []string{"a", "b", "c", "d", "eee"}

	unsupportedMessage = `unsupported request type`
)

func Test_Respond(t *testing.T) {
	options := WebServerOptions{
		Addr: listenAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method != "GET" {
				RespondWithText(w, 400, unsupportedMessage)
			} else if req.RequestURI == "/" || req.RequestURI == "/index.html" {
				RespondWithHtml(w, 200, indexData)
			} else if strings.HasSuffix(req.RequestURI, ".json") {
				RespondWithJson(w, 200, jsonData)
			} else if req.RequestURI == "/jsonstrings" {
				RespondWithJson(w, 200, dataStrings)
			} else {
				RespondWith(w, 404, "", `not found`)
			}
		}),
	}
	webServer := NewWebServer(options)
	if err := webServer.Start(); err != nil {
		t.Fatal(err)
	}
	// Test bad POST reqeust.
	url := fmt.Sprintf("http://%s/index.html", listenAddr)
	r0, err := http.Post(url, "application/yaml", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Error for POST to url=%s: %s", url, err)
	}
	if r0.StatusCode != 400 {
		t.Errorf("Expected status code=400 but instead found status code=%v for POST %s", r0.StatusCode, url)
	}
	foundContentType := r0.Header.Get("Content-Type")
	if foundContentType != MimePlain {
		t.Errorf("Expected content-type=%s but instead found content-type=%s for GET %s", MimePlain, foundContentType, url)
	}
	body, err := ioutil.ReadAll(r0.Body)
	if err != nil {
		t.Fatalf("Error reading body for POST to url=%s: %s", url, err)
	}
	if string(body) != unsupportedMessage {
		t.Errorf(`Expected body="%s" but instead found content="%s" for POST %s`, unsupportedMessage, string(body), url)
	}
	// Test GET / and /index.html.
	for _, page := range []string{"", "index.html"} {
		url = fmt.Sprintf("http://%s/%s", listenAddr, page)
		r1, err := http.Get(url)
		if err != nil {
			t.Fatalf("Error for GET from url=%s: %s", url, err)
		}
		if r1.StatusCode != 200 {
			t.Errorf("Expected status code=200 but instead found status code=%v for GET %s", r1.StatusCode, url)
		}
		foundContentType := r1.Header.Get("Content-Type")
		if foundContentType != MimeHtml {
			t.Errorf("Expected content-type=%s but instead found content-type=%s for GET %s", MimeHtml, foundContentType, url)
		}
		body, err := ioutil.ReadAll(r1.Body)
		if err != nil {
			t.Fatalf("Error reading body for GET from url=%s: %s", url, err)
		}
		if string(body) != string(indexData) {
			t.Errorf(`Expected body="%s" but instead found content="%s" for GET %s`, string(indexData), string(body), url)
		}
	}
	// Test GET /.*.json.
	for _, page := range []string{".json", "foo.json", "bar.json", "dir/foo.json", "dir1/dir2/dir3/bar.json"} {
		for _, prefix := range []string{"", "/", "//", "///"} {
			url = fmt.Sprintf("http://%s/%s%s", listenAddr, prefix, page)
			r2, err := http.Get(url)
			if err != nil {
				t.Fatalf("Error for GET from url=%s: %s", url, err)
			}
			if r2.StatusCode != 200 {
				t.Errorf("Expected status code=200 but instead found status code=%v for GET %s", r2.StatusCode, url)
			}
			foundContentType := r2.Header.Get("Content-Type")
			if foundContentType != MimeJson {
				t.Errorf("Expected content-type=%s but instead found content-type=%s for GET %s", MimeJson, foundContentType, url)
			}
			body, err := ioutil.ReadAll(r2.Body)
			if err != nil {
				t.Fatalf("Error reading body for GET from url=%s: %s", url, err)
			}
			if string(body) != string(jsonData) {
				t.Errorf(`Expected body="%s" but instead found content="%s" for GET %s`, string(jsonData), string(body), url)
			}
		}
	}
	// Test slice of strings JSON serialization.
	url = fmt.Sprintf("http://%s/jsonstrings", listenAddr)
	r3, err := http.Get(url)
	if err != nil {
		t.Fatalf("error for GET from url=%s: %s", url, err)
	}
	if r3.StatusCode != 200 {
		t.Errorf("Expected status code=200 but instead found status code=%v for GET %s", r3.StatusCode, url)
	}
	foundContentType = r3.Header.Get("Content-Type")
	if foundContentType != MimeJson {
		t.Errorf("Expected content-type=%s but instead found content-type=%s for GET %s", MimeJson, foundContentType, url)
	}
	body, err = ioutil.ReadAll(r3.Body)
	if err != nil {
		t.Fatalf("Error reading body for GET from url=%s: %s", url, err)
	}
	jsonified, err := json.Marshal(dataStrings)
	if err != nil {
		t.Fatalf("Error serializing `dataStrings' to JSON: %s", err)
	}
	if string(body) != string(jsonified) {
		t.Errorf(`Expected body="%s" but instead found body="%s" for GET %s`, string(jsonified), string(body), url)
	}
	if err := webServer.Stop(); err != nil {
		t.Fatal(err)
	}
}
