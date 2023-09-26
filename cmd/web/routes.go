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

	mux.Get("/test", handlers.Repo.TestSession)                           // Tests if the stack is stitched
	mux.Get("/", handlers.Repo.Login)                                     // login for a registered guy and oauth2 for a non-one
	mux.Get("/{loginAction}", handlers.Repo.ShopifyLogin)                 // api call for auth
	mux.Get("/turn_off_deal_list", handlers.Repo.TurnOffDealList)         // request to turn off deal list
	mux.Get("/turn_off_next_campaign", handlers.Repo.TurnOffNextCampaign) // turns off the campaign for next day only
	mux.Get("/campaign_activity", handlers.Repo.GetCampaignActivity)      // api to send active campaign activity
	mux.Get("/deallist_activity", handlers.Repo.GetDealListActivity)      // api to send overall deal list activity
	mux.Get("/trending_products", handlers.Repo.GetTrendingProducts)      // trending products list (from campaign_product)
	mux.Get("/otf_visitors", handlers.Repo.GetOtfVisitorData)             // series data for otf visitors (should come from the agg DB)
	mux.Get("/past_campaigns", handlers.Repo.GetAllCampaigns)             // list of daily campaigns for a specific store
	mux.Post("/config_discounts", handlers.Repo.ConfigureDiscounts)       // Configure discounts for a store
	mux.Post("/config_dl", handlers.Repo.ConfigureDealList)               // configure deal list properties for a store
	mux.Post("/update_profile", handlers.Repo.UpdateUserProfile)          // api to change user profile details
	mux.Post("/update_password", handlers.Repo.UpdatePassword)            // change dashboard password
	mux.Get("/if_otf", handlers.Repo.GetOtfUserInfo)                      // Pulls clickstream, aggregates in Postgres, and uses otf algo

	/* later for admin side login
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth) // and also check if acceslevel is top
		mux.Get("/dashboard", handlers.Repo.AdminDashboard) // goes to /admin/dashboard
	}) */

	return mux
}
