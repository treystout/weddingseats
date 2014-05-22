package weddingseats

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/context"

	"code.google.com/p/goauth2/oauth"

	"appengine"
	"appengine/datastore"
)

type User struct {
	FacebookID                string `json:"id"`
	FirstName                 string `json:"first_name"`
	LastName                  string `json:"last_name"`
	Gender                    string `json:"gender"`
	FacebookAccessToken       string
	FacebookAccessTokenExpiry time.Time
	TokenExpiry               time.Time
	DateCreated               time.Time
}

var ANONYMOUS *User

func init() {
	ANONYMOUS = &User{
		FacebookID:  "",
		FirstName:   "anonymous",
		LastName:    "",
		Gender:      "female",
		DateCreated: time.Unix(0, 0),
	}
}

func (u *User) Token() *oauth.Token {
	token := &oauth.Token{
		AccessToken: u.FacebookAccessToken,
		Expiry:      u.FacebookAccessTokenExpiry,
	}
	return token
}

func (u *User) Login(w http.ResponseWriter, r *http.Request) {
	// call this to register a known User with this user's session
	session, _ := SessionStore.Get(r, KeySessionCookieName)
	session.Values["user_key"] = u.FacebookID
	session.Save(r, w)
}

func GetUserFromSession(r *http.Request) (user *User) {
	ctx := appengine.NewContext(r)
	// first see if this user is in the global context for the request
	if rv := context.Get(r, KeyCurrentUser); rv != nil {
		ctx.Infof("user found in global context")
		return rv.(*User)
	}

	session, _ := SessionStore.Get(r, KeySessionCookieName)
	user_key, found := session.Values["user_key"]
	if !found {
		return ANONYMOUS
	}
	user, err := LocateUser(&ctx, session.Values["user_key"].(string))
	if err != nil {
		ctx.Debugf("session key found for user: %s but that user wasn't in the datastore!", user_key)
		return ANONYMOUS
	}
	// great we found them, put them in the global context
	context.Set(r, KeyCurrentUser, *user)
	return user
}

func (u *User) Key(ctx *appengine.Context) *datastore.Key {
	return datastore.NewKey(*ctx, "User", u.FacebookID, 0, nil)
}

func (u *User) Save(ctx *appengine.Context) error {
	u.DateCreated = time.Now()
	(*ctx).Debugf("saving user %+v", u)
	_, err := datastore.Put(*ctx, u.Key(ctx), u)
	if err != nil {
		return err
	}
	return nil
}

func LocateUser(ctx *appengine.Context, facebookID string) (u *User, err error) {
	u = new(User)
	u.FacebookID = facebookID
	err = datastore.Get(*ctx, u.Key(ctx), u)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, errors.New("not found!")
		} else {
			return nil, err
		}
	}
	return u, nil
}
