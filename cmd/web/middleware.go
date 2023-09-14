package main

import (
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
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		app.InfoLog.Println("Authentication successful")
		next.ServeHTTP(w, r)
	})
}

func AddTestStoreToSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		user.ID = 0
		user.Email = "tester@slaash.it"
		user.FirstName = "Slaash"
		user.LastName = "Tester"
		user.Store.ID = 1
		user.Store.ApiToken = "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"
		user.Store.Name = "spend-more-money.myshopify.com"

		session.Put(r.Context(), "user", user)
		app.InfoLog.Println("Test store added to the session")
		next.ServeHTTP(w, r)
	})
}
