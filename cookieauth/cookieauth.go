package cookieauth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/securecookie"
)

// CookieAuth provides secure pure cookie-based authentication capabilities to
// sign and set an encrypted userId in a cookie along with the timestamp of
// when the cookie was signed/created.
type CookieAuth struct {
	cookieName   string
	expireAfter  time.Duration
	secureCookie *securecookie.SecureCookie
}

const createdAtLayout = "2006-01-02 15:04:05 -0700 MST"

const NotPresentErrorString = "http: named cookie not present"

var Expired = errors.New("auth cookie expired")

// New creates a new instance of CookieAuth.
func New(hashKey []byte, blockKey []byte, cookieName string, expireAfter time.Duration) *CookieAuth {
	cookieAuth := &CookieAuth{
		cookieName:   cookieName,
		expireAfter:  expireAfter,
		secureCookie: securecookie.New(hashKey, blockKey),
	}
	return cookieAuth
}

// Read the cookie from a request and check for expiration.
func (cookieAuth *CookieAuth) Read(req *http.Request) (userId int64, err error) {
	var cookie *http.Cookie
	if cookie, err = req.Cookie(cookieAuth.cookieName); err != nil {
		if err.Error() == NotPresentErrorString {
			err = nil // Ignore this "error" and instead just return nil userId result.
		} else {
			err = fmt.Errorf("Cookieauth.Read error reading cookie: %s", err)
		}
		return
	}
	if len(cookie.Value) == 0 {
		return
	}
	value := make(map[string]string)
	if err = cookieAuth.secureCookie.Decode(cookieAuth.cookieName, cookie.Value, &value); err != nil {
		err = fmt.Errorf("Cookieauth.Read error decrypting cookie: %s", err)
		return
	}
	// Extract userId.
	userIdString, ok := value["userId"]
	if !ok {
		// No userId found in cookie value, return userId=0.
		log.WithField("action", "cookieauth.read").Infof("no userId found in value=%v", value)
		return
	}
	if userId, err = strconv.ParseInt(userIdString, 10, 64); err != nil {
		err = fmt.Errorf("Cookieauth.Read failed to parse userId=%v: %s", userIdString, err)
		return
	}
	// Extract and validate expiration timestamp.
	createdAtString, ok := value["createdAt"]
	if !ok {
		// No createdAt found in cookie value, return userId=0.
		log.WithField("action", "cookieauth.read").Infof("no createdAt found in value=%v", value)
		return
	}
	createdAt, err := time.Parse(createdAtLayout, createdAtString)
	log.WithField("action", "cookieauth.read").Debugf("createdAt=%v since=%v", createdAt, time.Since(createdAt))
	if time.Since(createdAt) > cookieAuth.expireAfter {
		// Expired.
		log.WithField("action", "cookieauth.read").Infof("discovered expired auth cookie createdAt=%s which more than %s ago", createdAt, cookieAuth.expireAfter)
		userId = 0 // Zero out userId.
		err = Expired
		return
	}
	log.WithField("action", "cookieauth.read").Infof("found userId=%v", value)
	return
}

// Set sets a userId cookie.  Should be applied only after successful authentication.
func (cookieAuth *CookieAuth) Set(userId int64, w http.ResponseWriter) error {
	value := map[string]string{
		"userId":    fmt.Sprintf("%v", userId),
		"createdAt": time.Now().Format(createdAtLayout),
	}
	encoded, err := cookieAuth.secureCookie.Encode(cookieAuth.cookieName, value)
	if err != nil {
		return fmt.Errorf("Cookieauth.Set: %s", err)
	}
	cookie := &http.Cookie{
		Name:  cookieAuth.cookieName,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
	return nil
}

// Clear removes the auth cookie.
func (cookieAuth *CookieAuth) Clear(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:  cookieAuth.cookieName,
		Value: "",
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}

// GenerateRandomKey is a convenience method which provides a pass-through to
// the securecookie.GenerateRandomKey function.
func GenerateRandomKey(length int) []byte {
	key := securecookie.GenerateRandomKey(length)
	return key
}
