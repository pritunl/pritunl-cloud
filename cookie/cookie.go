package cookie

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/pritunl/pritunl-cloud/session"
	"github.com/pritunl/pritunl-cloud/settings"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

var (
	AdminStore *sessions.CookieStore
	UserStore  *sessions.CookieStore
)

type Cookie struct {
	Id    bson.ObjectId
	store *sessions.Session
	w     http.ResponseWriter
	r     *http.Request
}

func (c *Cookie) Get(key string) string {
	valInf := c.store.Values[key]
	if valInf == nil {
		return ""
	}
	return valInf.(string)
}

func (c *Cookie) Set(key string, val string) {
	c.store.Values[key] = val
}

func (c *Cookie) GetSession(db *database.Database, r *http.Request,
	typ string) (sess *session.Session, err error) {

	sessId := c.Get("id")
	if sessId == "" {
		err = &errortypes.NotFoundError{
			errors.New("cookie: Session not found"),
		}
		return
	}

	sess, err = session.GetUpdate(db, sessId, r, typ)
	if err != nil {
		switch err.(type) {
		case *database.NotFoundError:
			err = &errortypes.NotFoundError{
				errors.New("cookie: Session not found"),
			}
		default:
			err = &errortypes.UnknownError{
				errors.Wrap(err, "cookie: Unknown session error"),
			}
		}
		return
	}

	return
}

func (c *Cookie) NewSession(db *database.Database, r *http.Request,
	id bson.ObjectId, remember bool, typ string) (
	sess *session.Session, err error) {

	sess, err = session.New(db, r, id, typ)
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "cookie: Unknown session error"),
		}
		return
	}

	c.Set("id", sess.Id)
	maxAge := 0

	if remember {
		maxAge = 15778500
	}

	c.store.Options.MaxAge = maxAge

	err = c.Save()
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "cookie: Unknown session error"),
		}
		return
	}

	return
}

func (c *Cookie) Remove(db *database.Database) (err error) {
	sessId := c.Get("id")
	if sessId == "" {
		return
	}

	err = session.Remove(db, sessId)
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "cookie: Unknown session error"),
		}
		return
	}

	c.Set("id", "")
	err = c.Save()
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err, "cookie: Unknown session error"),
		}
		return
	}

	return
}

func (c *Cookie) Save() (err error) {
	err = c.store.Save(c.r, c.w)
	return
}

func init() {
	module := requires.New("cookie")
	module.After("settings")

	module.Handler = func() (err error) {
		db := database.GetDatabase()
		defer db.Close()

		adminCookieAuthKey := settings.System.AdminCookieAuthKey
		adminCookieCryptoKey := settings.System.AdminCookieCryptoKey
		userCookieAuthKey := settings.System.UserCookieAuthKey
		userCookieCryptoKey := settings.System.UserCookieCryptoKey

		if len(adminCookieAuthKey) == 0 || len(adminCookieCryptoKey) == 0 {
			adminCookieAuthKey = securecookie.GenerateRandomKey(64)
			adminCookieCryptoKey = securecookie.GenerateRandomKey(32)
			settings.System.AdminCookieAuthKey = adminCookieAuthKey
			settings.System.AdminCookieCryptoKey = adminCookieCryptoKey

			fields := set.NewSet(
				"admin_cookie_auth_key",
				"admin_cookie_crypto_key",
			)

			err = settings.Commit(db, settings.System, fields)
			if err != nil {
				return
			}
		}

		if len(userCookieAuthKey) == 0 || len(userCookieCryptoKey) == 0 {
			userCookieAuthKey = securecookie.GenerateRandomKey(64)
			userCookieCryptoKey = securecookie.GenerateRandomKey(32)
			settings.System.UserCookieAuthKey = userCookieAuthKey
			settings.System.UserCookieCryptoKey = userCookieCryptoKey

			fields := set.NewSet(
				"user_cookie_auth_key",
				"user_cookie_crypto_key",
			)

			err = settings.Commit(db, settings.System, fields)
			if err != nil {
				return
			}
		}

		AdminStore = sessions.NewCookieStore(
			adminCookieAuthKey, adminCookieCryptoKey)
		AdminStore.Options.Secure = true
		AdminStore.Options.HttpOnly = true

		UserStore = sessions.NewCookieStore(
			userCookieAuthKey, userCookieCryptoKey)
		UserStore.Options.Secure = true
		UserStore.Options.HttpOnly = true

		return
	}
}
