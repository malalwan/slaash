package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/malalwan/slaash/internal/config"
	"github.com/malalwan/slaash/internal/driver"
	"github.com/malalwan/slaash/internal/models"
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
	User models.Store
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
		ClientID:     m.App.MyAppCreds[0],
		ClientSecret: m.App.MyAppCreds[1],
		RedirectURL:  m.App.RedirectURL,
		Scopes:       m.App.MyScopes,
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
	campaigns, err := m.DB.GetActiveCampaign(store)
	if err != nil {
		log.Fatal(err)
	}
	listOfProducts, err := m.DB.GetCampaignProducts(campaigns[0])
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
	var c models.Campaign
	c.Store = store
	action := chi.URLParam(r, "action")

	switch action {
	case "create":
		err := m.DB.CreateCampaign(c)
		_, _ = m.DB.CreateCampaignProducts(c)
		// send list of products for the form
		if err != nil {
			log.Fatal(err)
		}
	case "update":
		// Get campaignID from the request
		campaign, _ := m.DB.GetCampaignByID(c.CampaignID) // TBC
		_, err := m.DB.UpdateCampaignProducts(campaign)
		if err != nil {
			log.Fatal(err)
		}
	case "view":
		campagin, err := m.DB.GetCampaignByID(c.CampaignID)
		_, _ = m.DB.GetCampaignProducts(campagin)
		if err != nil {
			log.Fatal(err)
		}
		// send the list as a response
	case "list":
		_, _ = m.DB.ListAllCampaigns()
		// send back the fucking list
	}
}

// Majors renders the room page
func (m *Repository) SendSeriesData(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	_ = user.Store
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

}

// if login window and func for login post method is handled
// do this first
// m.app.Session.RenewToken(r.Context())
// then put the usr_id or smthing in the session
