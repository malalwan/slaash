package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/malalwan/slaash/internal/config"
	"github.com/malalwan/slaash/internal/driver"
	"github.com/malalwan/slaash/internal/helpers"
	"github.com/malalwan/slaash/internal/models"
	"github.com/malalwan/slaash/internal/render"
	"github.com/malalwan/slaash/internal/repository"
	"github.com/malalwan/slaash/internal/repository/dbrepo"
	"golang.org/x/oauth2"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App  *config.AppConfig
	DB   repository.DatabaseRepo
	User helpers.Store
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(m *Repository) {
	Repo = m
}

/*
ShopifyLogin handles user authentication and authorization
Prerequisites: None
Input: Store URL, Session Info, DB Record against User, Store
Output:
(1) User, Store exists : Ask for login/redirect to dashboard
(2) Authorize and retrieve access token, store against User, Store
(TODO)
*/
func (m *Repository) ShopifyLogin(w http.ResponseWriter, r *http.Request) {
	// Display a link to start the login process
	la := chi.URLParam(r, "loginAction")
	host := r.Host
	oauthConf := &oauth2.Config{
		ClientID:     m.App.MyApp.ID,
		ClientSecret: m.App.MyApp.Secret,
		RedirectURL:  m.App.MyApp.RedirectURL,
		Scopes:       m.App.MyApp.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s/admin/oauth/authorize", host),
			TokenURL: fmt.Sprintf("https://%s/admin/oauth/access_token", host),
		},
	}
	switch la {
	case "":
		loginURL := oauthConf.AuthCodeURL("", oauth2.AccessTypeOffline)
		fmt.Fprintf(w, `<a href="%s">Login with Shopify</a>`, loginURL)
	case "login":
		http.Redirect(w, r, oauthConf.AuthCodeURL("", oauth2.AccessTypeOffline), http.StatusFound)
	case "callback":
		// Exchange authorization code for access token
		code := r.URL.Query().Get("code")
		token, err := oauthConf.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "Error exchanging code for token", http.StatusInternalServerError)
			return
		}
		// Display the access token
		fmt.Fprintf(w, "Access Token: %s", token.AccessToken)
	}
}

/*
SendNoDeal sends no deal email for the current campaign
Prerequisites: User, Store must be active and there must be an active campaign
Input: Session Info : Users, Store
Output: Ack on DB Set and Email Queued
*/
func (m *Repository) SendNoDeal(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store
	campaign, err := models.GetActiveCampaign(store)
	if err != nil {
		log.Fatal(err)
	}
	listOfProducts, err := models.GetCampaignProducts(campaign)
	if err != nil {
		log.Fatal(err)
	}

	for _, prod := range listOfProducts {
		prod.DealDiscount = 0
		// push to DB
		// set campign in DB as inactive
	}
}

// TakeCampaignAction will CRUD campaigns for a store and display them
func (m *Repository) TakeCampaignAction(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store
	action := chi.URLParam(r, "action")

	switch action {
	case "create":
		campaign, err := models.CreateCampaign(store)
		listOfProducts, err := models.CreateCampaignProducts(campaign)
		// send list of products for the form
		if err != nil {
			log.Fatal(err)
		}
	case "update":
		// Get campaignID from the request
		campaign, err := models.GetCampaignByID(r.Response.Body) // TBC
		listOfProducts, err := models.UpdateCampaignProducts(campaign)
		if err != nil {
			log.Fatal(err)
		}
	case "view":
		campagin, err := models.GetCampaignByID(r.Response.Body)
		listOfProducts, err := models.GetCampaignProducts(campaign)
		if err != nil {
			log.Fatal(err)
		}
		// send the list as a response
	case "list":
		campaigns, err := models.ListAllCampaigns()
		// send back the fucking list
	}
}

// Majors renders the room page
func (m *Repository) SendSeriesData(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store
	column := chi.URLParam(r, "colId")
	time := chi.URLParam(r, "tStub")
	// startTime := (TODO)
	// endTime := (assing empty)
	if time[len(time)-1] == 'h' {
		// set prev time here
	} else if time[len(time)-1] == 'd' {
		// set prev date here
	}
	// connect to db and get ready
	switch column {
	case "gmv":
		// fetch group by purchases (campaign products)
	case "products":
		// fetch group by products (sold?) (campaign products)
	case "users":
		// fetch group by dlusers who gave email (buyer)
	case "discounts":
		// fetch group by discount amounts (campaign products)
	}
}

// Availability renders the search availability page
func (m *Repository) ShowCampaignStats(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}
