package authorizer

import (
	"net/http"

	"github.com/pritunl/pritunl-cloud/auth"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/signature"
)

func AuthorizeAdmin(db *database.Database, w http.ResponseWriter,
	r *http.Request) (authr *Authorizer, err error) {

	token := r.Header.Get("Auth-Token")
	sigStr := r.Header.Get("Auth-Signature")

	if token != "" && sigStr != "" {
		timestamp := r.Header.Get("Auth-Timestamp")
		nonce := r.Header.Get("Auth-Nonce")

		sig, e := signature.Parse(
			token,
			sigStr,
			timestamp,
			nonce,
			r.Method,
			r.URL.Path,
		)
		if e != nil {
			err = e
			return
		}

		err = sig.Validate(db)
		if err != nil {
			return
		}

		authr = &Authorizer{
			typ: Admin,
			sig: sig,
		}
	} else {
		cook, sess, e := auth.CookieSessionAdmin(db, w, r)
		if e != nil {
			err = e
			return
		}

		authr = &Authorizer{
			typ:  Admin,
			cook: cook,
			sess: sess,
		}
	}

	return
}

func AuthorizeUser(db *database.Database, w http.ResponseWriter,
	r *http.Request) (authr *Authorizer, err error) {

	token := r.Header.Get("Auth-Token")
	sigStr := r.Header.Get("Auth-Signature")

	if token != "" && sigStr != "" {
		timestamp := r.Header.Get("Auth-Timestamp")
		nonce := r.Header.Get("Auth-Nonce")

		sig, e := signature.Parse(
			token,
			sigStr,
			timestamp,
			nonce,
			r.Method,
			r.URL.Path,
		)
		if e != nil {
			err = e
			return
		}

		authr = &Authorizer{
			typ: User,
			sig: sig,
		}
		return
	}

	cook, sess, err := auth.CookieSessionUser(db, w, r)
	if err != nil {
		return
	}

	authr = &Authorizer{
		typ:  User,
		cook: cook,
		sess: sess,
	}

	return
}

func NewAdmin() (authr *Authorizer) {
	authr = &Authorizer{
		typ: Admin,
	}

	return
}

func NewUser() (authr *Authorizer) {
	authr = &Authorizer{
		typ: User,
	}

	return
}
