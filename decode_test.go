package web

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/gigawattio/testlib"
)

func TestDecodeJson(t *testing.T) {
	type InboundObject struct {
		Id          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var (
		src      = bytes.NewBufferString(`{"id": 999999950, "name": "J. Dawg", "description": "This is so informative!"}`)
		dst      = &InboundObject{}
		expected = &InboundObject{
			Id:          999999950,
			Name:        "J. Dawg",
			Description: "This is so informative!",
		}
	)

	if err := DecodeJson(src, dst); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(dst, expected) {
		t.Fatalf("Result did not match expected value\nActual=%+v\nExpected=%+v", dst, expected)
	}
}

func TestAllDecoders(t *testing.T) {
	type InboundObject struct {
		Id          int64    `json:"id" xml:"id" yaml:"id"`
		Name        string   `json:"name" xml:"name" yaml:"name"`
		Description string   `json:"description" xml:"description" yaml:"description"`
		Favorites   []string `json:"favorites" xml:"favorites" yaml:"favorites"`
	}

	var (
		inputFuncMap = map[string]func(r io.Reader, dst interface{}) error{
			`{
	"id": 9238418748311,
	"name": "J. Z",
	"description": "This is so informative!",
	"favorites": ["a", "b", "z"]
}`: DecodeJson,

			`<?xml version="1.0"?>
<inbound>
	<id>9238418748311</id>
	<name>J. Z</name>
	<description>This is so informative!</description>
	<favorites>a</favorites>
	<favorites>b</favorites>
	<favorites>z</favorites>
</inbound>`: DecodeXml,

			`
id: 9238418748311
name: J. Z
description: This is so informative!
favorites:
  - a
  - b
  - z`: DecodeYaml,
		}

		expected = InboundObject{
			Id:          9238418748311,
			Name:        "J. Z",
			Description: "This is so informative!",
			Favorites:   []string{"a", "b", "z"},
		}
	)

	for input, fn := range inputFuncMap {
		var (
			src = bytes.NewBufferString(input)
			dst = InboundObject{}
		)

		if err := fn(src, &dst); err != nil {
			t.Errorf("Problem decoding via func %s: %s", testlib.FullFunctionName(fn), err)
			continue
		}

		if !reflect.DeepEqual(dst, expected) {
			t.Errorf("Result did not match expected value via func %s\nActual=%+v\nExpected=%+v", testlib.FullFunctionName(fn), dst, expected)
		}
	}
}
