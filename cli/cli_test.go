package cli

import (
	// "bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/gigawattio/go-commons/pkg/errorlib"
	"github.com/gigawattio/go-commons/pkg/testlib"
	service "github.com/gigawattio/go-commons/pkg/web/cli/example/service"

	"github.com/parnurzeal/gorequest"
	cliv2 "gopkg.in/urfave/cli.v2"
)

func TestCli(t *testing.T) {
	options := Options{
		AppName:    testlib.CurrentRunningTest(),
		WebService: &service.MyWebService{},
		Args:       []string{"-b", "127.0.0.1:0"},
	}
	cli, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	cli.App.Action = func(ctx *cliv2.Context) error {
		if err := options.WebService.Start(ctx); err != nil {
			t.Fatal(err)
		}
		resp, body, errs := gorequest.New().Get(fmt.Sprintf("http://%s/", cli.WebService.Addr())).End()
		if err := errorlib.Merge(errs); err != nil {
			t.Error(err)
		}
		if expected := http.StatusOK; resp.StatusCode != expected {
			t.Errorf("Expected response status-code=%v but actual=%v", expected, resp.StatusCode)
		}
		if expected := "hello world"; body != expected {
			t.Errorf("Expected response body=%q but actual=%q", expected, body)
		}
		return nil
	}
	if err := cli.Main(); err != nil {
		t.Fatal(err)
	}
}

func TestCliAppNameError(t *testing.T) {
	options := Options{
		WebService: &service.MyWebService{},
		Args:       []string{"-b", "127.0.0.1:0"},
	}
	_, err := New(options)
	if expected := AppNameRequiredError; err != expected {
		t.Fatalf("Expected error=%v but actual=%s", expected, err)
	}
}

func TestCliWebServiceError(t *testing.T) {
	options := Options{
		AppName: testlib.CurrentRunningTest(),
		Args:    []string{"-b", "127.0.0.1:0"},
	}
	_, err := New(options)
	if expected := WebServiceRequiredError; err != expected {
		t.Fatalf("Expected error=%v but actual=%s", expected, err)
	}
}
