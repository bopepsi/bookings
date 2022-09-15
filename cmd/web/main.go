package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bopepsi/bookings/pkg/config"
	"github.com/bopepsi/bookings/pkg/handlers"
	"github.com/bopepsi/bookings/pkg/render"
)

const portNumber = ":8080"

// Setup app main package wide config
var app config.AppConfig

var session *scs.SessionManager

func main() {

	app.InProduction = false

	cache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}
	app.TemplateCache = cache
	app.UseCache = true

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true // exist after close window
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Setup handlers repo and render template cache
	handlers.SetupRepo(&app)
	render.SetupTmplCacheMap(&app)

	server := &http.Server{
		Addr:    ":8080",
		Handler: routes(&app),
	}
	log.Fatal(server.ListenAndServe())

}
