package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gigawattio/go-commons/pkg/errorlib"
	"github.com/gigawattio/go-commons/pkg/testlib"
	service "github.com/gigawattio/go-commons/pkg/web/cli/example/service"
	"github.com/gigawattio/go-commons/pkg/web/interfaces"

	"github.com/parnurzeal/gorequest"
	cliv2 "gopkg.in/urfave/cli.v2"
)

func simpleWebServiceProvider(ctx *cliv2.Context) (interfaces.WebService, error) {
	return service.New(ctx.String("bind")), nil
}

// genTestCliArgs prepends the current running tests name to the slice (in lieu
// of the binary filename) to form a slice whose contents mimic realistic
// command-line arguments.
func genTestCliArgs(args ...string) []string {
	args = append([]string{testlib.CurrentRunningTest()}, args...)
	return args
}

func TestCli(t *testing.T) {
	options := Options{
		AppName:            testlib.CurrentRunningTest(),
		WebServiceProvider: simpleWebServiceProvider,
		Args:               genTestCliArgs("--bind", "127.0.0.1:0"),
	}
	cli, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	cli.App.Action = func(ctx *cliv2.Context) error {
		webService, err := options.WebServiceProvider(ctx)
		if err != nil {
			return err
		}
		if err := webService.Start(); err != nil {
			t.Fatal(err)
		}
		resp, body, errs := gorequest.New().Get(fmt.Sprintf("http://%s/", webService.Addr())).End()
		if err := errorlib.Merge(errs); err != nil {
			t.Error(err)
		}
		if expected := http.StatusOK; resp.StatusCode != expected {
			t.Errorf("Expected response status-code=%v but actual=%v", expected, resp.StatusCode)
		}
		if expected := "hello world"; body != expected {
			t.Errorf("Expected response body=%q but actual=%q", expected, body)
		}
		if err := webService.Stop(); err != nil {
			t.Error(err)
		}
		return nil
	}
	if err := cli.Main(); err != nil {
		t.Fatal(err)
	}
}

func TestCliBindFlagWhenDefaultPortIsInUse(t *testing.T) {
	// Start on the default bind address:port.
	{
		defaultWebService := service.New(DefaultBindAddr)
		if err := defaultWebService.Start(); err != nil {
			t.Fatal(err)
		}
		resp, body, errs := gorequest.New().Get(fmt.Sprintf("http://%s/", defaultWebService.Addr())).End()
		if err := errorlib.Merge(errs); err != nil {
			t.Error(err)
		}
		if expected := http.StatusOK; resp.StatusCode != expected {
			t.Errorf("Expected response status-code=%v but actual=%v", expected, resp.StatusCode)
		}
		if expected := "hello world"; body != expected {
			t.Errorf("Expected response body=%q but actual=%q", expected, body)
		}
		defer func() {
			if err := defaultWebService.Stop(); err != nil {
				t.Fatal(err)
			}
		}()
	}

	options := Options{
		AppName:            testlib.CurrentRunningTest(),
		WebServiceProvider: simpleWebServiceProvider,
		Args:               genTestCliArgs("-b", "127.0.0.1:0"),
	}
	cli, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	cli.App.Action = func(ctx *cliv2.Context) error {
		webService, err := options.WebServiceProvider(ctx)
		if err != nil {
			return err
		}
		if err := webService.Start(); err != nil {
			t.Fatal(err)
		}
		resp, body, errs := gorequest.New().Get(fmt.Sprintf("http://%s/", webService.Addr())).End()
		if err := errorlib.Merge(errs); err != nil {
			t.Error(err)
		}
		if expected := http.StatusOK; resp.StatusCode != expected {
			t.Errorf("Expected response status-code=%v but actual=%v", expected, resp.StatusCode)
		}
		if expected := "hello world"; body != expected {
			t.Errorf("Expected response body=%q but actual=%q", expected, body)
		}
		if err := webService.Stop(); err != nil {
			t.Error(err)
		}
		return nil
	}
	if err := cli.Main(); err != nil {
		t.Fatal(err)
	}
}

func TestCliAppNameError(t *testing.T) {
	options := Options{
		WebServiceProvider: simpleWebServiceProvider,
		Args:               genTestCliArgs("-b", "127.0.0.1:0"),
	}
	_, err := New(options)
	if expected := AppNameRequiredError; err != expected {
		t.Fatalf("Expected error=%v but actual=%s", expected, err)
	}
}

func TestCliWebServiceProviderError(t *testing.T) {
	options := Options{
		AppName: testlib.CurrentRunningTest(),
		Args:    genTestCliArgs("-b", "127.0.0.1:0"),
	}
	_, err := New(options)
	if expected := WebServiceProviderRequiredError; err != expected {
		t.Fatalf("Expected error=%v but actual=%s", expected, err)
	}
}

func brokenWebServiceProvider(ctx *cliv2.Context) (interfaces.WebService, error) {
	return nil, nil
}

func TestCliBrokenWebServiceProvider(t *testing.T) {
	var (
		fakeStderr = &bytes.Buffer{}
		options    = Options{
			AppName:            testlib.CurrentRunningTest(),
			Usage:              "Fully automatic :)",
			Version:            "1024.2048.4096",
			Args:               genTestCliArgs("-b", "127.0.0.1:0"),
			WebServiceProvider: brokenWebServiceProvider,
			Stderr:             fakeStderr,      // Suppress os.Stderr output.
			Stdout:             &bytes.Buffer{}, // Suppress os.Stderr output.
			ExitOnError:        false,
		}
	)
	c, err := New(options)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan error, 2)
	go func() { ch <- c.Main() }()
	select {
	case actual := <-ch:
		if expected := NilWebServiceError; actual != expected {
			t.Errorf("Expected error=%v but actual=%s", expected, actual)
		}
		if expected, actual := NilWebServiceError.Error(), strings.Trim(fakeStderr.String(), "\n\t "); actual != expected {
			t.Errorf("Expected stderr to contain value=%q but actual=%q", expected, actual)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out after 100ms waiting for expected error")
	}
}

func TestCliOutputDefaults(t *testing.T) {
	options := Options{
		AppName:            testlib.CurrentRunningTest(),
		Args:               genTestCliArgs("-b", "127.0.0.1:0"),
		WebServiceProvider: simpleWebServiceProvider,
	}
	c, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	if c.App.Writer != os.Stdout {
		t.Errorf("Expected c.App.Writer == os.Stdout but it was set to something else instead; actual value=%T/%p", c.App.Writer, c.App.Writer)
	}
	if c.App.ErrWriter != os.Stderr {
		t.Errorf("Expected c.App.ErrWriter == os.Stderr but it was set to something else instead; actual value=%T/%p", c.App.ErrWriter, c.App.ErrWriter)
	}
}

func TestCliOutputOverrides(t *testing.T) {
	var (
		fakeStdout = &bytes.Buffer{}
		fakeStderr = &bytes.Buffer{}
		options    = Options{
			AppName:            testlib.CurrentRunningTest(),
			Args:               genTestCliArgs("-b", "127.0.0.1:0"),
			WebServiceProvider: simpleWebServiceProvider,
			Stdout:             fakeStdout,
			Stderr:             fakeStderr,
		}
	)
	c, err := New(options)
	if err != nil {
		t.Fatal(err)
	}
	if c.App.Writer != fakeStdout {
		t.Errorf("Expected c.App.Writer == fakeStdout (*bytes.Buffer) but it was set to something else instead; actual value=%T/%p", c.App.Writer, c.App.Writer)
	}
	if c.App.ErrWriter != fakeStderr {
		t.Errorf("Expected c.App.ErrWriter == fakeStderr (*bytes.Buffer) but it was set to something else instead; actual value=%T/%p", c.App.ErrWriter, c.App.ErrWriter)
	}
}
