package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func (m *Repository) TestSession(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store
	fmt.Printf("user.FirstName: %v\n", user.FirstName)
	fmt.Printf("user.LastName: %v\n", user.LastName)
	fmt.Printf("store.ApiToken: %v\n", store.ApiToken)
	fmt.Printf("store.Name: %v\n", store.Name)
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
	campaign, err := m.DB.GetActiveCampaign(store)
	if err != nil {
		log.Fatal(err)
	}
	listOfProducts, err := m.DB.GetCampaignProducts(campaign.CampaignID)
	if err != nil {
		log.Fatal(err)
	}

	for _, prod := range listOfProducts {
		prod.DealDiscount = 0
		// push to DB
		// set campign in DB as inactive
	}
}

func (m *Repository) SendTrendingProducts(w http.ResponseWriter, r *http.Request) {
	/* Initialize the function with the user and store context from the session */
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store

	/* DB API to fetch the most added products by end customers */
	list, err := m.DB.GetTopProductsByStore(store.ID)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	data := models.TopProducts{}

	for _, product := range list {
		/* Get product from shopify to fetch images etc
		We can pre store this info while initializing in the DB as well
		Since we get only 2 things from shopify: title and image
		*/
		p, err := store.GetProductById(product.ProductID)
		if err != nil {
			m.App.ErrorLog.Println(err)
		}
		var prod struct {
			ProductName  string
			ProductImage string
			Users        int
			Discount     struct {
				Value        int
				CurrencyType string
			}
			Gmv struct {
				Value        int
				CurrencyType string
			}
		}
		prod.ProductName = p.Title
		prod.ProductImage = p.Image.Src
		prod.Users = product.Deals
		prod.Discount.CurrencyType = "US Dollar"
		prod.Discount.Value = product.DealDiscount
		prod.Gmv.CurrencyType = "US Dollar"
		prod.Gmv.Value = product.Price - product.DealDiscount

		data.Products = append(data.Products, prod)
	}

	// Marshal the map into a JSON string
	jsonData, err := json.Marshal(data)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

/*
SendAggregateData returns the aggregate count gmv, roducts, and users
incorporated via deal list
*/
func (m *Repository) SendActiveCampaignAggregate(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store

	campaign, err := m.DB.GetActiveCampaign(store)
	if err != nil {
		fmt.Println(err)
	}
	data, err := m.DB.SelectFromCampaignById(int64(campaign.CampaignID), time.Now(),
		"SUM(price), COUNT(*), SUM(deals), SUM(dealdiscount*price/100)", "campaign_product", "campaignid = $1 and timestamp <= $2")
	if err != nil {
		m.App.ErrorLog.Println(err)
	}
	stats := models.AggregateStats{}
	stats.ActiveCampaignID = campaign.CampaignID
	stats.ActiveUsers.ActiveUsersInSession = data["users"]
	stats.Discount.Value = data["discounts"]
	stats.GmvActiveSession.CurrencyType = "US Dollar"
	stats.GmvActiveSession.Value = data["gmv"]
	stats.ProductsActiveSession.Products = data["products"]

	// Marshal the map into a JSON string
	jsonData, err := json.Marshal(stats)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

func (m *Repository) SendAggregateData(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		DuratonType  string `json:"durationType"`
		DealListType string `json:"dealListType"`
	}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		m.App.ErrorLog.Println(err)
		return
	}

	// typ := requestBody.DealListType
	duration := requestBody.DuratonType
	startTime := time.Now()
	switch duration {
	case "12hours":
		startTime = startTime.Add(-12 * time.Hour)
	case "24hours":
		startTime = startTime.Add(-24 * time.Hour)
	case "weekly":
		startTime = startTime.Add(-24 * 7 * time.Hour)
	case "monthly":
		startTime = startTime.Add(-24 * 30 * time.Hour)
	}

	/* SELECT SUM(price), COUNT(*), SUM(deals), SUM(dealdiscount*price/100)
	   FROM   campaign_product
	   WHERE  storeid = $1
	   AND	  timestamp >= $2 */

	data, err := m.DB.SelectFromCampaignById(int64(store.ID), startTime,
		"SUM(price), COUNT(*), SUM(deals), SUM(dealdiscount*price/100)", "campaign_product", "storeid = $1 and timestamp >= $2")
	if err != nil {
		m.App.ErrorLog.Println(err)
	}
	stats := models.AggregateForGraphs{}
	stats.DiscountSpends.Price = data["discounts"]
	stats.Users.Price = data["users"]
	stats.Gmv.Price = data["gmv"]
	stats.Products.Price = data["products"]
	/* SELECT DATE_TRUNC('hour', timestamp - interval '1 hour' * (EXTRACT(HOUR FROM timestamp) % 6)) AS interval,
	   SUM(price), COUNT(*),
	   SUM(deals), SUM(dealdiscount*price/100)
	   FROM   campaign_product
	   WHERE  storeid = $1
	   AND	  timestamp >= $2
	   GROUP BY interval
	   ORDER BY interval;
	*/
	// Marshal the map into a JSON string
	seriesData, err := m.DB.GetGroupSeriesData(int64(store.ID), startTime)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	stats.GmvData = seriesData[0]
	stats.ProductsData = seriesData[1]
	stats.UsersData = seriesData[2]
	stats.DiscountsData = seriesData[3]

	jsonData, err := json.Marshal(stats)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

func (m *Repository) SendOtfVisitorData(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		DurationType string `json:"durationType"`
		Type         string `json:"type"`
	}

	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		m.App.ErrorLog.Println(err)
		return
	}
	startTime := time.Now()
	switch requestBody.DurationType {
	case "12hours":
		startTime = startTime.Add(-12 * time.Hour)
	case "24hours":
		startTime = startTime.Add(-24 * time.Hour)
	case "weekly":
		startTime = startTime.Add(-24 * 7 * time.Hour)
	case "monthly":
		startTime = startTime.Add(-24 * 30 * time.Hour)
	}

	data, err := m.DB.GetAggregateOtfByDuration(startTime, requestBody.Type, store.ID)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	stats := models.OtfResponse{}

	stats.Otf = data

	jsonData, err := json.Marshal(stats)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

func (m *Repository) SendAllCampaigns(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	store := user.Store

	// data, err := m.DB.ListAllCampaigns(store.ID)

}

// if login window and func for login post method is handled
// do this first
// m.app.Session.RenewToken(r.Context())
// then put the user in the session
