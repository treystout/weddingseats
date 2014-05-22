package weddingseats

import (
	"net/http"
	"sync"

	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/sessions"
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

func init() {
	// locate and read our configuration
	err := ReadConfig("conf.json.prod")
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
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/tz", HandleGender)
	http.HandleFunc("/facebook_start", HandleFacebookStart)
	http.HandleFunc("/facebook_authorized", HandleFacebookAuthorized)
}
