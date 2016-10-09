package web

import (
	"fmt"
	"net/http"
	"testing"
)

func TestFsServer(t *testing.T) {
	fsServer := NewFsServer("127.0.0.1:0", ".")
	if err := fsServer.Start(); err != nil {
		t.Fatal(err)
	}
	response, err := http.Get(fmt.Sprintf("http://%s/", fsServer.Addr()))
	if err != nil {
		t.Errorf("Retrieving FsServer index page: %s", err)
	}
	if expected, actual := 200, response.StatusCode; actual != expected {
		t.Errorf("Expected response status-code=%v but actual=%v", expected, actual)
	}
	if err := fsServer.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestFsServerRestarts(t *testing.T) {
	fsServer := NewFsServer("127.0.0.1:0", ".")
	for i, _ := range [10]struct{}{} {
		if err := fsServer.Start(); err != nil {
			t.Fatalf("i=%v; Error starting: %s", err)
		}
		response, err := http.Get(fmt.Sprintf("http://%s/", fsServer.Addr()))
		if err != nil {
			t.Errorf("i=%v; Retrieving FsServer index page: %s", i, err)
		}
		if expected, actual := 200, response.StatusCode; actual != expected {
			t.Errorf("i=%v; Expected response status-code=%v but actual=%v", i, expected, actual)
		}
		if err := fsServer.Stop(); err != nil {
			t.Fatalf("i=%v; Error stopping: %s", err)
		}
	}
}
