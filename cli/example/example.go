package main

import (
	"github.com/gigawattio/go-commons/pkg/web/cli"
	"github.com/gigawattio/go-commons/pkg/web/cli/example/service"
)

func main() {
	options := cli.Options{
		AppName:     "Example App",
		WebService:  &service.MyWebService{},
		Args:        []string{"-b", "127.0.0.1:0"},
		ExitOnError: true,
	}
	cli, err := cli.New(options)
	if err != nil {
		panic(err)
	}
	cli.Main()
}
