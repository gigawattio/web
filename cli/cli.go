package cli

// cli provides a rails-like system for running command-line interface
// web-services without requiring repetitive boilerplate.

import (
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/gigawattio/go-commons/pkg/upstart"
	"github.com/gigawattio/go-commons/pkg/web/interfaces"

	cliv2 "gopkg.in/urfave/cli.v2"
)

var (
	DefaultBindAddr    = "127.0.0.1:8080"
	DefaultServiceUser = os.Getenv("USER")
)

// Options provider for Cli struct.
type Options struct {
	AppName            string
	Usage              string
	Version            string
	Flags              []cliv2.Flag
	Stdout             io.Writer
	Stderr             io.Writer
	WebServiceProvider interfaces.WebServiceProvider
	Args               []string
	ExitOnError        bool // Exit on non-nil error during invocation of `Main()`.
}

// Cli provides a command-line-interface in-a-box for web-services.
type Cli struct {
	App                *cliv2.App
	WebServiceProvider interfaces.WebServiceProvider
	Args               []string
	Install            bool   // NB: Flag variable.
	Uninstall          bool   // NB: Flag variable.
	ServiceUser        string // NB: Flag variable.
	BindAddr           string // NB: Flag variable.
	ExitOnError        bool   // true triggers os.exit upon error from Main().
	initialized        bool
}

func New(options Options) (*Cli, error) {
	cli := &Cli{
		App: &cliv2.App{
			Name:      options.AppName,
			Version:   options.Version,
			Usage:     options.Usage,
			Flags:     options.Flags,
			Writer:    options.Stdout,
			ErrWriter: options.Stderr,
		},
		WebServiceProvider: options.WebServiceProvider,
		Args:               options.Args,
		ExitOnError:        options.ExitOnError,
	}
	if err := cli.Init(); err != nil {
		return nil, err
	}
	return cli, nil
}

// Init performs validation and when possible auto-populates absent fields with
// the `os` package equivalent (e.g. if `cli.App.Writer == nil` then it is set to
// `os.Stderr`) or other default value.
func (cli *Cli) Init() error {
	if len(cli.App.Name) == 0 {
		return AppNameRequiredError
	}
	if cli.WebServiceProvider == nil {
		return WebServiceProviderRequiredError
	}

	// Don't change or check anything if already initialized.
	if cli.initialized {
		return nil
	}

	// Replace empty fields with default values where possible.
	if len(cli.ServiceUser) == 0 {
		cli.ServiceUser = DefaultServiceUser
	}
	if len(cli.BindAddr) == 0 {
		cli.BindAddr = DefaultBindAddr
	}

	// Auto-populate empty fields with os.* equivalents.
	if cli.Args == nil {
		cli.Args = os.Args
	}
	if cli.App.ErrWriter == nil {
		cli.App.Writer = os.Stdout
	}
	if cli.App.ErrWriter == nil {
		cli.App.ErrWriter = os.Stderr
	}

	// App flags and action.
	if cli.App.Flags == nil {
		cli.App.Flags = []cliv2.Flag{}
	}
	cli.App.Flags = append(
		cli.App.Flags,
		[]cliv2.Flag{
			&cliv2.BoolFlag{
				Name:        "install",
				Usage:       fmt.Sprintf("Install %s as a system service", cli.App.Name),
				Destination: &cli.Install,
			},
			&cliv2.BoolFlag{
				Name:        "uninstall",
				Usage:       fmt.Sprintf("Uninstall %s as a system service", cli.App.Name),
				Destination: &cli.Uninstall,
			},
			&cliv2.StringFlag{
				Name:        "user",
				Aliases:     []string{"u"},
				Usage:       fmt.Sprintf("Specifies the user to run the %s system service as", cli.App.Name),
				Value:       cli.ServiceUser,
				Destination: &cli.ServiceUser,
			},
			&cliv2.StringFlag{
				Name:        "bind",
				Aliases:     []string{"b"},
				Usage:       "Set the web-server bind-address and (optionally) port",
				Value:       cli.BindAddr,
				Destination: &cli.BindAddr,
			},
		}...,
	)
	// Setup default action
	cli.App.Action = cli.DefaultAction

	cli.initialized = true // Mark as initialized.

	return nil
}

// DefaultAction can be overriden by setting cli.App.Action after invoking
// `Init()`.
func (cli *Cli) DefaultAction(ctx *cliv2.Context) error {
	if err := cli.ServiceManagementHandler(); err != nil {
		return err
	}
	if err := cli.RunWeb(ctx); err != nil {
		return err
	}
	return nil
}

// ServiceManagementHandler performs system service management updates.
func (cli *Cli) ServiceManagementHandler() error {
	if cli.Uninstall || cli.Install {
		if cli.Uninstall {
			config := upstart.DefaultConfig(cli.App.Name)
			if err := upstart.UninstallService(config); err != nil {
				return err
			}
		}
		if cli.Install {
			config := upstart.DefaultConfig(cli.App.Name)
			config.User = cli.ServiceUser
			if err := upstart.InstallService(config); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cli *Cli) RunWeb(ctx *cliv2.Context) error {
	if cli.WebServiceProvider == nil {
		return WebServiceProviderRequiredError
	}
	webService, err := cli.WebServiceProvider(ctx)
	if err != nil {
		return err
	}
	if webService == nil {
		return NilWebServiceError
	}
	if err := webService.Start(); err != nil {
		return err
	}
	fmt.Fprintf(cli.App.Writer, "Successfully started web service on addr=%v\n", webService.Addr())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig // Wait for ^C signal.
	fmt.Fprintln(cli.App.ErrWriter, "\nInterrupt signal detected, shutting down..")

	if err := webService.Stop(); err != nil {
		return err
	}

	return nil
}

func (cli *Cli) Main() error {
	// Temporarily disable cliv2 os exiter and redirect ErrWriter to the one for
	// this app.
	var (
		backupOsExiter  = cliv2.OsExiter
		backupErrWriter = cliv2.ErrWriter
	)
	cliv2.OsExiter = func(_ int) {}
	cliv2.ErrWriter = cli.App.ErrWriter
	defer func() {
		// Restore backed up values.
		cliv2.OsExiter = backupOsExiter
		cliv2.ErrWriter = backupErrWriter
	}()

	if err := cli.App.Run(cli.Args); err != nil {
		if cli.ExitOnError {
			ErrorExit(cli.App.ErrWriter, err, 1)
		} else {
			return err
		}
	}
	return nil
}

func ErrorExit(w io.Writer, err error, statusCode int) {
	fmt.Fprintf(w, "error: %s\n", err)
	fmt.Fprintf(w, "exiting with status-code=%v\n", statusCode)
	os.Exit(statusCode)
}
