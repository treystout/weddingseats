package weddingseats

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/goauth2/oauth"

	"appengine"
	"appengine/urlfetch"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	// net/http appears to default to your most permissive route ("/" in this case)
	// so defend against favicon and friends from also hitting this route
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	c := appengine.NewContext(r)
	other_user, err := LocateUser(&c, "798890670134774")
	if err != nil {
		c.Warningf("couldn't find user! %s", err.Error())
	}
	c.Debugf("got this user back %+v", other_user)
	user := GetUserFromSession(r)

	w.Header().Set("Content-Type", "text/html")
	session, _ := SessionStore.Get(r, "somedude")
	counter, found := session.Values["counter"]
	log.Printf("counter from session: %+v", counter)
	if !found {
		log.Printf("counter not found in session")
		counter = 0
	}
	session.Values["counter"] = counter.(int) + 1
	log.Printf("counter inserted into session as: %+v", counter)
	session.Save(r, w)
	fmt.Fprintf(w, "<h1>Oh hai there: %s</h1>", user.FirstName)
	fmt.Fprintf(w, "counter: %d<br>", counter)
	fmt.Fprintf(w, "<a href=\"/facebook_start\">login with facebook</a>")
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

	// fetch their /me info
	resp, err := transport.Client().Get("https://graph.facebook.com/me")
	if err != nil {
		c.Errorf("couldn't get /me (%s)", err.Error())
	}
	defer resp.Body.Close()
	/*
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.Errorf("couldn't read body (%s)", err.Error())
		}
	*/

	// try to unmarshall the json
	user := new(User)
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber() // to get our int64's out
	err = decoder.Decode(user)
	if err != nil {
		c.Errorf("couldn't decode json from facebook! %s", err.Error())
	}
	c.Debugf("decoded /me as user: %+v", user)
	user.Save(&c)
	user.Login(w, r)

	http.Redirect(w, r, "/", http.StatusFound)
	//fmt.Fprintf(w, "facebook /me: %s", string(body))
}
