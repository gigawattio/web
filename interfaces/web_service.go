package interfaces

import (
	"net"

	cliv2 "gopkg.in/urfave/cli.v2"
)

type WebService interface {
	Start() error
	Stop() error
	Addr() net.Addr
}

type WebServiceProvider func(ctx *cliv2.Context) (WebService, error)
