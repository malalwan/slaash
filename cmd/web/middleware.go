package main

import (
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/malalwan/slaash/internal/helpers"
	"github.com/malalwan/slaash/internal/models"
)

/* NoSurf is the csrf protection middleware */
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

/* SessionLoad loads and saves session data for current request */
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Log in first!")
			//http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			fmt.Fprintf(w, "Authentication Failed, Retry!")
			return
		}
		/* add a check for the sexy otf api, if it's one of the stores, let it come in */
		app.InfoLog.Println("Authentication successful")
		next.ServeHTTP(w, r)
	})
}

func AddTestStoreToSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.Users
		user.Email = "tester@slaash.it"
		user.FirstName = "Slaash"
		user.LastName = "Tester"
		user.Store = 1
		user.AccessLevel = 1
		user.Misc = "Test user for Slaash"
		// user.Password = "test"
		//user.Store.ApiToken = "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"
		//user.Store.Name = "spend-more-money.myshopify.com"

		session.Put(r.Context(), "user", user)
		app.InfoLog.Println("Test store added to the session")
		next.ServeHTTP(w, r)
	})
}
