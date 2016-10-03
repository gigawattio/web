package auth

import (
	"encoding/base64"
	"net/http"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	// Provide a minimal test implementation.
	authOpts := AuthOptions{
		Realm:    "Restricted",
		User:     "test-user",
		Password: "plain-text-password",
	}

	b := &basicAuth{
		opts: authOpts,
	}

	r := &http.Request{}
	r.Method = "GET"

	// Provide auth data, but no Authorization header.
	if b.authenticate(r) != false {
		t.Fatal("No Authorization header supplied.")
	}

	// Initialise the map for HTTP headers.
	r.Header = http.Header(make(map[string][]string))

	// Set a malformed/bad header.
	r.Header.Set("Authorization", "    Basic")
	if b.authenticate(r) != false {
		t.Fatal("Malformed Authorization header supplied.")
	}

	// Test correct credentials.
	auth := base64.StdEncoding.EncodeToString([]byte(b.opts.User + ":" + b.opts.Password))
	r.Header.Set("Authorization", "Basic "+auth)
	if b.authenticate(r) != true {
		t.Fatal("Failed on correct credentials")
	}
}

func TestBasicAuthAuthFunc(t *testing.T) {
	secret := "foobar123"
	authOpts := AuthOptions{
		Realm: "Protected",
		AuthFunc: func(username, password string) bool {
			if password == secret {
				return true
			}
			return false
		},
	}

	b := basicAuth{
		opts: authOpts,
	}

	r := &http.Request{}
	r.Method = "GET"

	// Provide auth data, but no Authorization header.
	if b.authenticate(r) != false {
		t.Fatal("No Authorization header supplied.")
	}

	// Initialise the map for HTTP headers
	r.Header = http.Header(make(map[string][]string))

	// Set a malformed/bad header
	r.Header.Set("Authorization", "    Basic")
	if b.authenticate(r) != false {
		t.Fatal("Malformed Authorization header supplied.")
	}

	{
		// Test incorrect credentials.
		auth := base64.StdEncoding.EncodeToString([]byte("any-name-is-fine:" + secret + "z"))
		r.Header.Set("Authorization", "Basic "+auth)
		if b.authenticate(r) != false {
			t.Fatal("Succeeded on incorrect credentials")
		}
	}

	{
		// Test correct credentials.
		auth := base64.StdEncoding.EncodeToString([]byte("any-name-is-fine:" + secret))
		r.Header.Set("Authorization", "Basic "+auth)
		if b.authenticate(r) != true {
			t.Fatal("Failed on correct credentials")
		}
	}
}
