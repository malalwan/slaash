package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/malalwan/slaash/pkg/config"
	"github.com/malalwan/slaash/pkg/handlers"
	"github.com/malalwan/slaash/pkg/render"
)

var app config.AppConfig
var session *scs.SessionManager

// Genesis Function
func main() {
	// change this to true in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)

	render.NewTemplates(&app)
	handlers.NewHandlers(repo)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}
