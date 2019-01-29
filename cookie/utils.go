package cookie

import (
	"net/http"

	"github.com/dropbox/godropbox/errors"
	"github.com/gorilla/securecookie"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

func GetAdmin(w http.ResponseWriter, r *http.Request) (
	cook *Cookie, err error) {

	store, err := AdminStore.New(r, "pritunl-cloud-admin")
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err.(securecookie.MultiError)[0],
				"cookie: Unknown cookie error"),
		}
		return
	}

	cook = &Cookie{
		store: store,
		w:     w,
		r:     r,
	}

	return
}

func NewAdmin(w http.ResponseWriter, r *http.Request) (cook *Cookie) {
	store, _ := AdminStore.New(r, "pritunl-cloud-admin")

	cook = &Cookie{
		store: store,
		w:     w,
		r:     r,
	}

	return
}

func CleanAdmin(w http.ResponseWriter, r *http.Request) {
	cook := &http.Cookie{
		Name:     "pritunl-cloud-admin",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cook)

	return
}

func GetUser(w http.ResponseWriter, r *http.Request) (
	cook *Cookie, err error) {

	store, err := UserStore.New(r, "pritunl-cloud-user")
	if err != nil {
		err = &errortypes.UnknownError{
			errors.Wrap(err.(securecookie.MultiError)[0],
				"cookie: Unknown cookie error"),
		}
		return
	}

	cook = &Cookie{
		store: store,
		w:     w,
		r:     r,
	}

	return
}

func NewUser(w http.ResponseWriter, r *http.Request) (cook *Cookie) {
	store, _ := UserStore.New(r, "pritunl-cloud-user")

	cook = &Cookie{
		store: store,
		w:     w,
		r:     r,
	}

	return
}

func CleanUser(w http.ResponseWriter, r *http.Request) {
	cook := &http.Cookie{
		Name:     "pritunl-cloud-user",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cook)

	return
}
