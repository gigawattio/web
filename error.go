package web

import (
	"fmt"
)

func JsonError(detail interface{}) Json {
	switch detail.(type) {
	case error:
		detail = detail.(error).Error()
	case string:
	default:
		detail = fmt.Sprint(detail)
	}
	log.ExtraCalldepth = 1
	log.Error("JsonError: detail=%v\n", detail)
	log.ExtraCalldepth = 0
	return Json{"error": detail}
}

func JsonErrorf(format string, args ...interface{}) Json {
	return JsonError(fmt.Sprintf(format, args))
}
