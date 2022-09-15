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

	return router
}
