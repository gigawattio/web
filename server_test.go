package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"testing"
)

const testAddr = "127.0.0.1:0"

func stopper(server *WebServer, t *testing.T) {
	if err := server.Stop(); err != nil {
		t.Fatalf("error stopping WebServer=%+v: %s", *server, err)
	}
	// Use `nc` to double check that the socket is fully closed.
	// This should fail because the socker is supposed to be finished.
	out, err := exec.Command("nc", append([]string{"-v", "-w", "1"}, strings.Split(server.Options.Addr, ":")...)...).CombinedOutput()
	if err == nil {
		t.Fatalf("Server socket still open for %+v at %s: %s\n----\nnc output:\n%s", *server, server.Options.Addr, err, string(out))
	}
	// o, _ := exec.Command("curl", testAddr).CombinedOutput()
	// t.Logf("%s\n", string(o))
}

func Test_StaticWebServer1(t *testing.T) {
	content := []byte("<!DOCTYPE html><html><head><title>Test page</title></head><body>hi, this is a test page.</body></html>")
	contentType := "text/html"
	server := NewStaticWebServer(WebServerOptions{Addr: testAddr}, content, http.StatusOK, map[string]string{"Content-Type": contentType})
	defer stopper(server, t)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	response, err := http.Get(fmt.Sprintf("http://%s/", server.Addr()))
	if err != nil {
		t.Fatal(err)
	}
	// Verify response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code=%v but instead got status code=%v", http.StatusOK, response.StatusCode)
	}
	// Verify response content-type header.
	foundContentType := response.Header.Get("Content-Type")
	if foundContentType != contentType {
		t.Errorf("Expected Content-Type header=%s but instead header=%s", contentType, foundContentType)
	}
	// Verify body content.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(content) {
		t.Errorf(`Expected body to be "%s" but instead found "%s"`, string(content), string(body))
	}
}

// Test_StaticWebServer2 ensures that webservers can successfully be brought up
// and down and up on the same address.
func Test_StaticWebServer2(t *testing.T) {
	content := []byte(`{"a":"json-document"}`)
	contentType := "application/json"
	server := NewStaticWebServer(WebServerOptions{Addr: testAddr}, content, http.StatusOK, map[string]string{"Content-Type": contentType})
	defer stopper(server, t)
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	response, err := http.Get(fmt.Sprintf("http://%s/", server.Addr()))
	// Invoke again to workaround weird issue where during the first request the
	// handler previously defined in Test_StaticWebServer1 is invoked before
	// ours.
	response, err = http.Get(fmt.Sprintf("http://%s/", server.Addr()))
	if err != nil {
		t.Fatal(err)
	}
	// Verify response status code.
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code=%v but instead got status code=%v", http.StatusOK, response.StatusCode)
	}
	// Verify response content-type header.
	foundContentType := response.Header.Get("Content-Type")
	if foundContentType != contentType {
		t.Errorf("Expected Content-Type header=%s but instead header=%s", contentType, foundContentType)
	}
	// Verify body content.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(content) {
		t.Errorf(`Expected body to be "%s" but instead found "%s"`, string(content), string(body))
	}
}

func Test_BaseUrl(t *testing.T) {
	server := NewWebServer(WebServerOptions{Addr: testAddr})
	if expected, actual := fmt.Sprintf("http://%v", server.Addr()), server.BaseUrl(); actual != expected {
		t.Errorf(`Expected BaseUrl="%s" but instead found "%s"`, expected, actual)
	}
}
