package weddingseats

import (
	"log"
	"net/http"
	"sync"

	"code.google.com/p/goauth2/oauth"
)

// global vars
var (
	config       *Configuration
	configLock   = new(sync.RWMutex) // so we can hot-reload config
	FACEBOOK_CFG = new(oauth.Config)
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	// locate and read our configuration
	err := ReadConfig("conf.json")
	check(err)

	FACEBOOK_CFG.ClientId = config.Facebook.ClientId
	FACEBOOK_CFG.ClientSecret = config.Facebook.ClientSecret
	FACEBOOK_CFG.AuthURL = config.Facebook.AuthURL
	FACEBOOK_CFG.TokenURL = config.Facebook.TokenURL
	FACEBOOK_CFG.RedirectURL = config.Facebook.RedirectURL

	log.Printf("conf was read: %+v", config)

	// setup URL handlers
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/facebook_start", HandleFacebookStart)
	http.HandleFunc("/facebook_authorized", HandleFacebookAuthorized)
}
