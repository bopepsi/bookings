package main

import (
	"net/http"

	"github.com/bopepsi/bookings/pkg/config"
	"github.com/bopepsi/bookings/pkg/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(WriteToConsole)
	router.Use(NoSurf)
	router.Use(LoadSession)

	router.Get("/", handlers.Repo.Home)
	router.Get("/about", handlers.Repo.About)
	router.Get("/index", handlers.Repo.Index)

	router.Get("/generals-quarters", handlers.Repo.Generals)
	router.Get("/majors-suite", handlers.Repo.Majors)
	router.Get("/search-availability", handlers.Repo.Availability)
	router.Get("/make-reservation", handlers.Repo.Reservation)
	router.Get("/contact", handlers.Repo.Contact)

	// serve static files
	fileServer := http.FileServer(http.Dir("static"))
	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return router
}
