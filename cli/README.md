# cli package notes

## Example usage

For a minimal usage example see the following file set:

* [example/example.go](example/example.go)
* [example/service/my_web_service.go](example/service/my_web_service.go)

## Run the example

    cd $GOPATH/src/github.com/gigawattio/go-commons/pkg/web/cli/example
    go run example.go

Output:

    Successfully started web service on addr=127.0.0.1:8080
    ^C
    Interrupt signal detected, shutting down..

## Additional notes

A few variables in the package `gopkg.in/urface/cli.v2` _do_ get _temporarily_ overridden during the invocation of `Cli.Main()` until the function is done running.  The overrides are:

* cliv2.OsExiter is _temporarily_ set to a NOP function.
* cliv2.ErrWriter is _temporarily_ set to the value of `Cli.App.ErrWriter`.