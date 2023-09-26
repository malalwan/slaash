package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
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
var postgresCreds driver.DBCreds
var clickhouseCreds driver.DBCreds

/* the main function */
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	app.InfoLog.Printf("Staring application on port %s", portNumber)

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

/*
	Function to set up all global entities:

1. AppConfig
2. PostgresDB Config
3. Clickhouse Config
*/
func run() (*driver.DB, error) {
	/* what am I going to put in the session? */
	gob.Register(models.Users{})

	/* to pick different DBs for test and prod and secure cookies */
	app.InProduction = false

	if app.InProduction {
		app.MyAppCreds = []string{"7f4b95c01d4764f01cb658adfad31108", "54bb5ac2cbf15a6d6bdb8bdeafec00f6"}
		app.MyScopes = []string{"dd", "bb"} // to be edited
		app.RedirectURL = "dashboard.slaash.it"
	} else {
		app.MyAppCreds = []string{"5e5ce46a1dfdf20f90f07016293f3838", "045f0b1fc37793af68f5bce04c9e2b63"}
		app.MyScopes = []string{"dd", "bb"}
		app.RedirectURL = "dashboard.slaash.it"
	}

	/* initializing loggers */
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	/* session to monitor logged in user */
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	/* push the session to global app config for easy access */
	app.Session = session

	/* Init postgres creds */
	postgresCreds.Host = "pg-slaash-slaash-01.aivencloud.com"
	postgresCreds.Port = 19236
	postgresCreds.Username = "avnadmin"
	postgresCreds.Password = "AVNS__o7qCttikmfMMABdM7J"
	app.InfoLog.Println("Connecting to Database")
	var dbname string
	if app.InProduction {
		dbname = "dashboard"
	} else {
		dbname = "defaultdb"
	}

	connString := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s",
		postgresCreds.Host, postgresCreds.Port, dbname,
		postgresCreds.Username, postgresCreds.Password)
	db, err := driver.ConnectSQL(connString)
	if err != nil {
		log.Fatal("Cannot connect to postgres database! Dying...")
	}
	app.InfoLog.Println("Connected to postgres database!")

	/* Init clickhouse creds */
	clickhouseCreds.Host = "clickhouse-slaash-01-slaash-01.aivencloud.com"
	clickhouseCreds.Port = 19237
	clickhouseCreds.Username = "avnadmin"
	clickhouseCreds.Password = "AVNS_vqSF_mM4PoCemxvdn7H"

	clickhouseString := fmt.Sprintf("https://%s:%d?username=%s&password=%s&secure",
		clickhouseCreds.Host, clickhouseCreds.Port,
		clickhouseCreds.Username, clickhouseCreds.Password)
	clickhouse, err := driver.ConnectClickhouse(clickhouseString)
	if err != nil {
		log.Fatal("Cannot connect to clickhouse database! Dying...")
	}
	app.InfoLog.Println("Connected to clickhouse database!")

	/* Set up globals for all packages */
	repo := handlers.NewRepo(&app, db, clickhouse)
	handlers.NewHandlers(repo)
	helpers.NewHelpers(&app)
	models.NewShopifyFunctions(&app)

	return db, nil
}
