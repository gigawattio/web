package web

import (
	"fmt"

	"github.com/facebookgo/stack"
	log "github.com/sirupsen/logrus"
)

func JsonError(detail interface{}) Json {
	switch detail.(type) {
	case error:
		detail = detail.(error).Error()
	case string:
	default:
		detail = fmt.Sprint(detail)
	}
	log.Errorf("%v: JsonError: detail=%v\n", stack.Caller(1), detail)
	return Json{"error": detail}
}

func JsonErrorf(format string, args ...interface{}) Json {
	return JsonError(fmt.Sprintf(format, args))
}
