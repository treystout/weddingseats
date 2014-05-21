package weddingseats

import (
	"encoding/json"
	"fmt"
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
	//c := appengine.NewContext(r)
	user := GetUserFromSession(r)

	//w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Oh hai there: %s</h1>", user.FirstName)
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
	//c.Debugf("looks like you got authed via facebook")

	// Since we're on GAE we can't use the default oauth transport's client,
	// so we're injecting the GAE urlfetcher instead
	transport := &oauth.Transport{
		Config:    FACEBOOK_CFG,
		Transport: &urlfetch.Transport{Context: c},
	}

	//
	code := r.FormValue("code")
	//c.Debugf("fb code: >>>%s<<<", code)
	token, err := transport.Exchange(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error exchanging Facebook token (%s)", err.Error()), http.StatusInternalServerError)
		return
	}
	//c.Debugf("facebook exchanged token: >>>%s<<<", token)

	// fetch their /me info
	resp, err := transport.Client().Get("https://graph.facebook.com/me")
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching Facebook info (%s)", err.Error()), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// try to unmarshall the json
	user := new(User)
	user.FacebookAccessToken = token.AccessToken
	user.FacebookAccessTokenExpiry = token.Expiry
	decoder := json.NewDecoder(resp.Body)
	//decoder.UseNumber() // to get our int64's out
	err = decoder.Decode(user)
	if err != nil {
		c.Errorf("couldn't decode json from facebook! %s", err.Error())
	}
	c.Debugf("decoded /me as user: %+v", user)
	err = user.Save(&c)
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't save user! %s", err.Error()), http.StatusInternalServerError)
		return
	}
	user.Login(w, r)

	http.Redirect(w, r, "/", http.StatusFound)
}

func HandleGender(w http.ResponseWriter, r *http.Request) {
	type Foo struct {
		Gender string `json:"gender"`
	}
	var foo Foo
	foo.Gender = "??"

	user := GetUserFromSession(r)
	if user.FacebookID != "" {
		c := appengine.NewContext(r)
		transport := &oauth.Transport{
			Config:    FACEBOOK_CFG,
			Token:     user.Token(),
			Transport: &urlfetch.Transport{Context: c},
		}
		resp, err := transport.Client().Get("https://graph.facebook.com/v2.0/me?fields=gender")
		if err != nil {
			http.Error(w, fmt.Sprintf("error fetching Facebook info (%s)", err.Error()), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&foo)
		if err != nil {
			c.Errorf("couldn't decode json from facebook! %s", err.Error())
		}
	}
	fmt.Fprintf(w, "gender from fb is %s", foo.Gender)
}
