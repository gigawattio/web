package web

import (
	"fmt"
	"net/http"
	"reflect"

	log "github.com/Sirupsen/logrus"
)

// TODO(jet) Add some form of logging control or toggle to this package,
// particularly for control over the RespondWith* functions.

type Json map[string]interface{}

// RespondWith tries to detect what kind of data is being sent and serializes it
// when appropriate, e.g. transmitting a struct, ptr, etc.
func RespondWith(w http.ResponseWriter, statusCode int, contentType string, data interface{}) (n int, err error) {
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(statusCode)
	var b []byte
INFER:
	switch data.(type) {
	case []byte:
		goto WRITE_RESPONSE
	case Json,
		[]Json,
		string,
		[]string,
		bool,
		int,
		[]int,
		int64,
		[]int64,
		map[string]string,
		map[string]int,
		map[string]int64,
		map[string]interface{}:
		goto MARSHAL

	default:
		v := reflect.ValueOf(data)
		switch v.Kind() {
		case reflect.Ptr:
			if v.IsNil() {
				err = fmt.Errorf("RespondWith %q: data must not be `nil'", contentType)
				return
			}
			data = v.Elem().Interface()
			goto INFER

		case reflect.Struct:
			goto MARSHAL

		case reflect.Slice:
			switch v.Slice(0, 0).Kind() {
			case reflect.Uint8: // []byte
				goto WRITE_RESPONSE

			default:
				goto MARSHAL // Hope for a slice of something serializable.
			}

		case reflect.Map:
			t := reflect.TypeOf(data)
			switch t.Key().String() {
			case "string":
				goto MARSHAL

			default:
				err = fmt.Errorf("RespondWith %q: automatic encoding failed for type=%T", contentType, data)
				return
			}

		default:
			err = fmt.Errorf("RespondWith %q: automatic encoding failed for type=%T", contentType, data)
			return
		}
	}
MARSHAL:
	{
		m, mErr := MarshallerFor(contentType)
		if mErr != nil {
			err = fmt.Errorf("RespondWith %q: %s", contentType, mErr)
			return
		}
		data, err = m(data)
		if err != nil {
			err = fmt.Errorf("RespondWith %q: marshalling error: %s", contentType, err)
			return
		}
	}
WRITE_RESPONSE:
	var ok bool
	b, ok = data.([]byte)
	if !ok {
		err = fmt.Errorf("RespondWith %q: failed to cast data to []byte, data=%v", contentType, data)
		return
	}
	n, err = w.Write(b)
	if err != nil {
		err = fmt.Errorf("RespondWith %q: error writing response: %s", contentType, err)
		return
	}
	return
}

func RespondWithJson(w http.ResponseWriter, statusCode int, data interface{}) (int, error) {
	return interceptErrors(RespondWith(w, statusCode, MimeJson, data))
}
func RespondWithXml(w http.ResponseWriter, statusCode int, data interface{}) (int, error) {
	return interceptErrors(RespondWith(w, statusCode, MimeXml, data))
}
func RespondWithYaml(w http.ResponseWriter, statusCode int, data interface{}) (int, error) {
	return interceptErrors(RespondWith(w, statusCode, MimeYaml2, data))
}
func RespondWithHtml(w http.ResponseWriter, statusCode int, data interface{}) (int, error) {
	return interceptErrors(RespondWith(w, statusCode, MimeHtml, data))
}
func RespondWithText(w http.ResponseWriter, statusCode int, data interface{}) (int, error) {
	return interceptErrors(RespondWith(w, statusCode, MimePlain, data))
}

func interceptErrors(statusCode int, err error) (int, error) {
	if err != nil {
		log.Errorf("Intercepted low-level HTTP response error during transmission: %s", err)
	}
	return statusCode, err
}
