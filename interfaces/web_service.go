package interfaces

import (
	"net"

	cliv2 "gopkg.in/urfave/cli.v2"
)

type WebService interface {
	Start(ctx *cliv2.Context) error
	Stop() error
	Addr() net.Addr
}
