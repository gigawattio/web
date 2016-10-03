package cookieauth

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/gigawattio/go-commons/pkg/web"
)

const (
	bindAddr   = "localhost:8585"
	cookieName = "a-cookie-name"
)

var (
	hashKey     = GenerateRandomKey(64)
	blockKey    = GenerateRandomKey(32)
	expireAfter = 1 * time.Second
	userId      = int64(99)
)

func Test_CookieAuth(t *testing.T) {
	cookieAuth := New(hashKey, blockKey, cookieName, expireAfter)
	server := web.NewWebServer(web.WebServerOptions{
		Addr: bindAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if err := cookieAuth.Set(userId, w); err != nil {
				t.Fatalf("`cookieAuth.Set' failed: %s", err)
			}
		}),
	})
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %s", err)
	}
	defer server.Stop()
	response, err := http.Get(fmt.Sprintf("http://%s/testing", bindAddr))
	if err != nil {
		t.Fatalf("HTTP client request failed: %s", err)
	}
	numCookies := len(response.Cookies())
	if numCookies != 1 {
		t.Fatalf("Expected 1 single cookie to be in the response, but instead found count=%v", numCookies)
	}
	cookie := response.Cookies()[0]
	if !regexp.MustCompile("^" + cookieName + "=").MatchString(cookie.String()) {
		t.Errorf(`Expected cookie string to start with "%s", but actual contents="%s"`, cookieName, cookie.String())
	}
	// Ensure CookieAuth.Read(req) with no cookie returns userId=0 and error=nil.
	request1, err := http.NewRequest("get", fmt.Sprintf("http://%s/", bindAddr), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Failed to create *http.Request #1 for testing: %s", err)
	}
	userId0, err := cookieAuth.Read(request1)
	if err != nil {
		t.Errorf("Expected err=nil for CookieAuth.Read(request1) when no cookies were attached, but instead found err=%v", err)
	}
	if userId0 != 0 {
		t.Errorf("Expected userId=0 for CookieAuth.Read(request1) when no cookies were attached, but instead found userId0=%v", userId0)
	}
	// Valid (non-expired) cookie scenario.
	request1.AddCookie(cookie)
	readUserId1, err := cookieAuth.Read(request1)
	if err != nil {
		t.Fatalf("Expected cookieAuth to successfully read the request, but got an error: %s", err)
	}
	if readUserId1 != userId {
		t.Errorf("Expected read userId=%v but instead found readUserId1=%v", userId, readUserId1)
	}
	// Expired cookie scenario.
	time.Sleep(expireAfter) // Ensure we wait until after the cookie is expired.
	request2, err := http.NewRequest("get", fmt.Sprintf("http://%s/", bindAddr), &bytes.Buffer{})
	request2.AddCookie(cookie)
	if err != nil {
		t.Fatalf("Failed to create *http.Request #2 for testing: %s", err)
	}
	readUserId2, err := cookieAuth.Read(request2)
	if err != Expired {
		t.Errorf(`Expected cookieAuth to throw the expiration error="%v", but got err="%v"`, Expired, err)
	}
	if readUserId2 != 0 {
		t.Errorf("Expected userId value to be zero'd out, but instead found readUserId2=%v", readUserId2)
	}
}
