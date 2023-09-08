package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/malalwan/slaash/internal/config"
	"github.com/malalwan/slaash/internal/driver"
	"github.com/malalwan/slaash/internal/handlers"
	"github.com/malalwan/slaash/internal/helpers"
	"github.com/malalwan/slaash/internal/models"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

/* the main function */
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, error) {
	/* what am I going to put in the session?
	will be used to fetch store context for the dashboard
	and display profile (editable?) */
	gob.Register(models.User{})

	/* to pick different DBs for test and prod */
	app.InProduction = false
	app.MyAppCreds = []string{"7f4b95c01d4764f01cb658adfad31108", "54bb5ac2cbf15a6d6bdb8bdeafec00f6"}
	app.MyScopes = []string{"dd", "bb"}
	app.RedirectURL = "dashboard.slaash.it"

	/* initialize my loggers
	Will be needed to show info and error logs across the code */
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	/* the session here is to monitor the
	users who have logged into the dashboard
	Secure cookies will hash out so won't be used for development */
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	/* push the session to global app config for easy access */
	app.Session = session

	/* connect to slaash database */
	log.Println("Connecting to database...")
	var dbname string
	if app.InProduction {
		dbname = "defaultdb"
	} else {
		dbname = "dashboard"
	}
	connString := fmt.Sprintf("host=pg-slaash-slaash-01.aivencloud.com port=19236 dbname=%s user=avnadmin password=AVNS__o7qCttikmfMMABdM7J", dbname)
	db, err := driver.ConnectSQL(connString)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}

	log.Println("Connected to database!")

	/* Repo is a wrapper over appconfig. It stores DB info
	over the global appconfig for request handling */
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	helpers.NewHelpers(&app)
	models.NewShopifyFunctions(&app)

	return db, nil
}
