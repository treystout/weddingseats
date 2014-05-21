package weddingseats

import (
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

type User struct {
	FacebookID  string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Gender      string `json:"gender"`
	DateCreated time.Time
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

func (u *User) Login(w http.ResponseWriter, r *http.Request) {
	// call this to register a known User with this user's session
	session, _ := SessionStore.Get(r, "session")
	session.Values["user_key"] = u.FacebookID
	session.Save(r, w)
}

func GetUserFromSession(r *http.Request) (user *User) {
	session, _ := SessionStore.Get(r, "session")
	user_key, found := session.Values["user_key"]
	if !found {
		return ANONYMOUS
	}
	ctx := appengine.NewContext(r)
	user, err := LocateUser(&ctx, session.Values["user_key"].(string))
	if err != nil {
		ctx.Debugf("session key found for user: %s but that user wasn't in the datastore!", user_key)
		return ANONYMOUS
	}
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
