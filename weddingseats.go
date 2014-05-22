package weddingseats

import (
	"net/http"
	"sync"

	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"

	"appengine"
)

// global vars
var (
	config       *Configuration
	configLock   = new(sync.RWMutex) // so we can hot-reload config
	FACEBOOK_CFG = new(oauth.Config)
	SessionStore *sessions.CookieStore
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func PreRequest(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("in pre request")
	context.Set(r, KeyCurrentUser, GetUserFromSession(r))
}
func PostRequest(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("in post request")
}

func WrapHandler(wrapped http.HandlerFunc) http.HandlerFunc {
	// wraps a given Handler with PreRequest and PostRequest
	return func(w http.ResponseWriter, r *http.Request) {
		PreRequest(w, r)
		// now wrap the wrapped call with a context-clearing handler
		clean_wrapped := context.ClearHandler(wrapped)
		clean_wrapped.ServeHTTP(w, r)
		// finally do the post request
		PostRequest(w, r)
	}
}

func init() {
	// locate and read our configuration
	//err := ReadConfig("conf.json.prod")
	err := ReadConfig("conf.json.testing")
	check(err)

	FACEBOOK_CFG.ClientId = config.Facebook.ClientId
	FACEBOOK_CFG.ClientSecret = config.Facebook.ClientSecret
	FACEBOOK_CFG.AuthURL = config.Facebook.AuthURL
	FACEBOOK_CFG.TokenURL = config.Facebook.TokenURL
	FACEBOOK_CFG.RedirectURL = config.Facebook.RedirectURL
	FACEBOOK_CFG.Scope = config.Facebook.Scope

	// setup session storage
	SessionStore = sessions.NewCookieStore([]byte(config.CookieSecret))
	SessionStore.Options = &sessions.Options{
		Secure: false,
	}

	// setup URL handlers
	http.Handle("/", WrapHandler(HandleIndex))
	http.Handle("/tz", WrapHandler(HandleGender))
	http.HandleFunc("/logout", HandleLogout)
	http.HandleFunc("/facebook_start", HandleFacebookStart)
	http.HandleFunc("/facebook_authorized", HandleFacebookAuthorized)
}
