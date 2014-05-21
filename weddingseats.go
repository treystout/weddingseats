package weddingseats

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"code.google.com/p/goauth2/oauth"

	"appengine"
	"appengine/urlfetch"
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

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<a href=\"/facebook_start\">login with facebook</a>")
}

func init() {
	err := ReadConfig("conf.json")
	check(err)

	FACEBOOK_CFG.ClientId = config.Facebook.ClientId
	FACEBOOK_CFG.ClientSecret = config.Facebook.ClientSecret
	FACEBOOK_CFG.AuthURL = config.Facebook.AuthURL
	FACEBOOK_CFG.TokenURL = config.Facebook.TokenURL
	FACEBOOK_CFG.RedirectURL = config.Facebook.RedirectURL

	log.Printf("conf was read: %+v", config)
	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/facebook_start", facebook_start)
	http.HandleFunc("/facebook_authorized", facebook_authorized)
}

func facebook_start(w http.ResponseWriter, r *http.Request) {
	url := FACEBOOK_CFG.AuthCodeURL("")
	c := appengine.NewContext(r)
	c.Infof("redirecting to %s", url)
	http.Redirect(w, r, url, http.StatusFound)
}

func facebook_authorized(w http.ResponseWriter, r *http.Request) {
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

func OLD_main() {
	log.Printf("wedding seats!\n")
	fp, err := os.Open("guests.csv")
	check(err)
	defer fp.Close()

	reader := csv.NewReader(fp)
	check(err)
	reader.TrailingComma = true
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	names := make([]string, 0, 200)

	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		names = append(names, fields[0])
	}

	for _, n := range names {
		log.Printf(">%+v<\n", n)
	}
}
