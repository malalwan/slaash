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
	if !app.InProduction {
		mux.Use(AddTestStoreToSession)
	}

	mux.Get("/{loginAction}", handlers.Repo.ShopifyLogin)                  // may not need this if auth is taken care in the front end
	mux.Get("/no_deal", handlers.Repo.SendNoDeal)                          // request to handle no deal
	mux.Get("/dl", handlers.Repo.SendAggregateData)                        // Aggregates and sends time series data for a store (from buyer table)
	mux.Get("/test", handlers.Repo.TestSession)                            // Tests if the stack is stitched
	mux.Get("/trending", handlers.Repo.SendTrendingProducts)               // trending products list (from campaign_product)
	mux.Get("/active_campaign", handlers.Repo.SendActiveCampaignAggregate) // Aggregated data for active campaign (should come from buyer table)
	mux.Get("/otf_data", handlers.Repo.SendOtfVisitorData)                 // series data for otf visitors (should come from the agg DB)
	mux.Get("/campaigns", handlers.Repo.SendAllCampaigns)                  // list of daily campaigns for a specific store

	/* // later for admin side login
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth) // and also check if acceslevel is top
		mux.Get("/dashboard", handlers.Repo.AdminDashboard) // goes to /admin/dashboard
	}) */

	return mux
}
