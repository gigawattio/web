package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Marshaller func(v interface{}) ([]byte, error)

func MarshallerFor(contentType string) (m Marshaller, err error) {
	switch contentType {
	case MimeHtml, MimePlain:
		m = marshalText
	case MimeJson, MimeJavascript:
		m = json.Marshal
	case MimeXml, MimeXml2:
		m = xml.Marshal
	case MimeYaml, MimeYaml2:
		m = yaml.Marshal
	default:
		err = fmt.Errorf("unable to encode unrecognized contentType=%s", contentType)
		return
	}
	return
}

func marshalText(value interface{}) ([]byte, error) {
	switch value.(type) {
	case string:
		s, _ := value.(string) // TODO: Cleanup/revisit
		return []byte(s), nil
	case []byte:
		b, _ := value.([]byte) // TODO: Cleanup/revisit
		return b, nil
	}
	return []byte{}, fmt.Errorf("unable to transform value=%T to []byte", value)
}

func DecodeJson(src io.Reader, value interface{}) error {
	decoder := json.NewDecoder(src)
	if err := decoder.Decode(value); err != nil {
		return err
	}
	return nil
}
func DecodeXml(src io.Reader, value interface{}) error {
	decoder := xml.NewDecoder(src)
	if err := decoder.Decode(value); err != nil {
		return err
	}
	return nil
}
func DecodeYaml(src io.Reader, value interface{}) error {
	in, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(in, value); err != nil {
		return err
	}
	return nil
}
