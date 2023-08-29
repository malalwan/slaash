package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/malalwan/slaash/internal/config"
	"github.com/malalwan/slaash/internal/handlers"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/{loginAction}", handlers.Repo.ShopifyLogin)
	mux.Get("/noDeal", handlers.Repo.SendNoDeal)
	mux.Get("/campAction/{action}", handlers.Repo.TakeCampaignAction)
	mux.Get("/dlCurves/{colId}/{tStub}", handlers.Repo.SendSeriesData)
	mux.Get("/campStore/{action}", handlers.Repo.ShowCampaignStats)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
