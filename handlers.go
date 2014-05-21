package weddingseats

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"code.google.com/p/goauth2/oauth"

	"appengine"
	"appengine/urlfetch"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<a href=\"/facebook_start\">login with facebook</a>")
}

func HandleFacebookStart(w http.ResponseWriter, r *http.Request) {
	url := FACEBOOK_CFG.AuthCodeURL("")
	c := appengine.NewContext(r)
	c.Infof("redirecting to %s", url)
	http.Redirect(w, r, url, http.StatusFound)
}

func HandleFacebookAuthorized(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Debugf("looks like you got authed via facebook")

	transport := &oauth.Transport{
		Config:    FACEBOOK_CFG,
		Transport: &urlfetch.Transport{Context: c},
	}
	code := r.FormValue("code")
	c.Debugf("fb code: >>>%s<<<", code)
	token, err := transport.Exchange(code)
	if err != nil {
		c.Errorf("token could not be exchanged: %s", err.Error())
	}
	// TODO: put the token in a cache and log the user in

	c.Debugf("facebook exchanged token: >>>%s<<<", token)

	// fetch their me info
	resp, err := transport.Client().Get("https://graph.facebook.com/me")
	if err != nil {
		c.Errorf("couldn't get /me (%s)", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Errorf("couldn't read body (%s)", err.Error())
	}

	fmt.Fprintf(w, "facebook /me: %s", string(body))
}
