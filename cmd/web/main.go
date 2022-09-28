package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bopepsi/bookings/internal/config"
	"github.com/bopepsi/bookings/internal/driver"
	"github.com/bopepsi/bookings/internal/handlers"
	"github.com/bopepsi/bookings/internal/helpers"
	"github.com/bopepsi/bookings/internal/models"
	"github.com/bopepsi/bookings/internal/render"
)

const portNumber = ":8080"

// var app config.AppConfig
var app config.AppConfig
var session *scs.SessionManager

var infoLog *log.Logger
var errorLog *log.Logger

// main is the main function
func main() {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Restriction{})
	gob.Register(models.RoomRestriction{})
	gob.Register(models.Room{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(app.MailChan)

	fmt.Println("Starting msg listener")
	listenForMail()

	// change this to true when in production
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=bopepsi password=")
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	log.Println("Connected to db")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	srv := &http.Server{
		Addr: portNumber,
		// Handler: routes(&app),
		Handler: routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	// r := routes()
	// log.Fatal(http.ListenAndServe(":8080", r))
}
