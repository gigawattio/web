package main

import (
	"github.com/gigawattio/go-commons/pkg/web/cli"
	"github.com/gigawattio/go-commons/pkg/web/cli/example/service"
	"github.com/gigawattio/go-commons/pkg/web/interfaces"

	cliv2 "gopkg.in/urfave/cli.v2"
)

func webServiceProvider(ctx *cliv2.Context) (interfaces.WebService, error) {
	webService := service.New(ctx.String("bind"))
	return webService, nil
}

func main() {
	options := cli.Options{
		AppName:            "Example App",
		Usage:              "More info here",
		WebServiceProvider: webServiceProvider,
		ExitOnError:        true,
	}
	c, err := cli.New(options)
	if err != nil {
		panic(err)
	}
	c.Main()
}
