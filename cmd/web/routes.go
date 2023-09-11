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
	//mux.Use(Auth)

	mux.Get("/{loginAction}", handlers.Repo.ShopifyLogin)              // may not need this if auth is taken care in the front end
	mux.Get("/noDeal", handlers.Repo.SendNoDeal)                       // request to handle no deal
	mux.Get("/campAction/{action}", handlers.Repo.TakeCampaignAction)  // CRUD a campaign
	mux.Get("/dl/{colId}/{tStub}", handlers.Repo.SendAggregateData)    // Aggregates and sends time series data for count
	mux.Get("/dlCurves/{colId}/{tStub}", handlers.Repo.SendSeriesData) // Sends an array of datapoints for curves
	mux.Get("/campStore/{action}", handlers.Repo.ShowCampaignStats)    // Campaign Statistics

	/* // later for admin side login
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth) // and also check if acceslevel is top
		mux.Get("/dashboard", handlers.Repo.AdminDashboard) // goes to /admin/dashboard
	}) */

	return mux
}
