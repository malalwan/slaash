package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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
	listOfProducts, err := m.DB.GetCampaignProducts(campaigns[0].CampaignID)
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
	/* create a campaign, add that id to all store products with unassigned campaignID with default discounts
	once that is done, return the list of products with the new campaignid.
	updation with edit those products
	deletion is basically no discount, will set discount as 0 for all those products, then return the value */
	// user := m.App.Session.Get(r.Context(), "user").(models.User)
	// store := user.Store
	store := models.Store{ID: 1} //stub
	var c models.Campaign
	c.CampaignID = 1 // stub
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
		_, _ = m.DB.GetCampaignProducts(campagin.CampaignID)
		if err != nil {
			log.Fatal(err)
		}
		// send the list as a response
	case "list":
		data, err := m.DB.ListAllCampaigns(store.ID)
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Println(len(data))
		fmt.Fprintf(w, "%s\n", jsonData)
	}
}

/*
SendAggregateData returns the aggregate count gmv, roducts, and users
incorporated via deal list
*/
func (m *Repository) SendAggregateData(w http.ResponseWriter, r *http.Request) {
	//user := m.App.Session.Get(r.Context(), "user").(models.User)
	//store := user.Store
	store := models.Store{ID: 1} //stub
	column := chi.URLParam(r, "colId")
	t := chi.URLParam(r, "tStub")
	startTime := time.Now()
	i, err := strconv.Atoi(t[:len(t)-1])
	if err != nil {
		log.Fatal(err)
	}
	if t[len(t)-1] == 'h' {
		startTime = startTime.Add(-1 * time.Duration(i) * time.Hour)
	} else if t[len(t)-1] == 'd' {
		startTime = startTime.Add(-24 * time.Duration(i) * time.Hour)
	}

	switch column {
	case "gmv":
		/* SELECT SUM(price)
		   FROM   campaign_product
		   WHERE  storeid = $1
		   AND	  timestamp >= $2 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "SUM(price)", "campaign_product", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	case "products":
		/* SELECT SUM(deals)
		   FROM   campaign_product
		   WHERE  storeid = $1
		   AND	  timestamp >= $2 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "SUM(deals)", "campaign_product", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	case "users":
		/* SELECT COUNT(*)
		   FROM   buyer
		   WHERE  storeid = $1
		   AND    timestamp >= $2
		   AND	  LENGTH(TRIM(email)) <> 0 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "COUNT(*)", "buyer", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	case "discounts":
		/* SELECT SUM(dealdiscount)
		   FROM   campaign_product
		   WHERE  stroreid = $1
		   AND	  timestamp >= $2 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "SUM(dealdiscount*price/100)", "campaign_product", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	}
}

func (m *Repository) SendSeriesData(w http.ResponseWriter, r *http.Request) {
	//user := m.App.Session.Get(r.Context(), "user").(models.User)
	//store := user.Store
	store := models.Store{ID: 1} //stub
	column := chi.URLParam(r, "colId")
	t := chi.URLParam(r, "tStub")
	startTime := time.Now()
	i, err := strconv.Atoi(t[:len(t)-1])
	if err != nil {
		log.Fatal(err)
	}
	if t[len(t)-1] == 'h' {
		startTime = startTime.Add(-1 * time.Duration(i) * time.Hour)
	} else if t[len(t)-1] == 'd' {
		startTime = startTime.Add(-24 * time.Duration(i) * time.Hour)
	}

	switch column {
	case "gmv":
		/* SELECT price
		   FROM   campaign_product
		   WHERE  stroreid = $1
		   AND	  timestamp >= $2 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "price, timestamp", "campaign_product", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	case "products":
		/* SELECT deals, timestamp
		   FROM   campaign_product
		   WHERE  stroreid = $1
		   AND	  timestamp >= $2 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "deals, timestamp", "campaign_product", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	case "users":
		/* SELECT timestamp
		   FROM   buyer
		   WHERE  storeid = $1
		   AND    timestamp >= $2
		   AND	  LENGTH(TRIM(email)) <> 0 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "anonymousid, timestamp", "buyer", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	case "discounts":
		/* SELECT dealdiscount
		   FROM   campaign_product
		   WHERE  stroreid = $1
		   AND	  timestamp >= $2 */
		data, err := m.DB.SelectFromCampaignByStore(store.ID, startTime, "dealdiscount*price/100, timestamp", "campaign_product", "storeid = $1 AND timestamp >= $2")
		if err != nil {
			fmt.Println(err)
		}
		// Marshal the map to a JSON string
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Print the JSON string using fmt.Fprintf
		fmt.Fprintf(w, "%s\n", jsonData)
	}
}

// Availability renders the search availability page
func (m *Repository) ShowCampaignStats(w http.ResponseWriter, r *http.Request) {

}

// if login window and func for login post method is handled
// do this first
// m.app.Session.RenewToken(r.Context())
// then put the user in the session
