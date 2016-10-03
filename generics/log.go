package generics

import (
	"github.com/op/go-logging"
)

const PACKAGE = "generics"

var (
	log  = logging.MustGetLogger(PACKAGE)
	logX = logging.MustGetLogger(PACKAGE) // Extra call-depth logger.
)

func init() {
	logX.ExtraCalldepth = 3
}

func SetExtraCallDepth(x int) {
	logX.ExtraCalldepth = x
}
