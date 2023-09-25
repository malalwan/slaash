package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/malalwan/slaash/internal/config"
	"github.com/malalwan/slaash/internal/driver"
	"github.com/malalwan/slaash/internal/helpers"
	"github.com/malalwan/slaash/internal/models"
	"github.com/malalwan/slaash/internal/repository"
	"github.com/malalwan/slaash/internal/repository/dbrepo"
	"golang.org/x/oauth2"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App        *config.AppConfig
	DB         repository.DatabaseRepo
	Clickhouse repository.ClickhouseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB, clickhouse *driver.DB) *Repository {
	return &Repository{
		App:        a,
		DB:         dbrepo.NewPostgresRepo(db.SQL, a),
		Clickhouse: dbrepo.NewClickhouseRepo(clickhouse.SQL, a),
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
func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
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

	loginURL := oauthConf.AuthCodeURL("", oauth2.AccessTypeOffline)
	fmt.Fprintf(w, `<a href="%s">Login with Shopify</a>`, loginURL)

}

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

func (m *Repository) TurnOffDealList(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store
	err := m.DB.StopDealList(storeid)
	if err != nil {
		m.App.ErrorLog.Println("Deal list turn off failed!")
		helpers.ServerError(w, err)
	}
}

func (m *Repository) TurnOffNextCampaign(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store
	err := m.DB.SetTurnOffTime(storeid)
	if err != nil {
		m.App.ErrorLog.Println("Failed to set turn off time for campaign")
		helpers.ServerError(w, err)
	}
}

func (m *Repository) GetCampaignActivity(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store
	endTime, err := m.DB.GetCampignEndTime(storeid) // call to store
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch Campaign timers")
		helpers.ServerError(w, err)
	}
	money, err := m.DB.GetAggFromCheckout(storeid) // call to checkout
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch GMV and Discount data")
		helpers.ServerError(w, err)
	}
	stats, err := m.DB.GetAggFromVisitor(storeid) // call to visitor
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch visitor data")
		helpers.ServerError(w, err)
	}
	data := models.CampaignActivity{}
	store, err := m.DB.GetStoreByID(storeid)
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch Store info from ID")
		helpers.ServerError(w, err)
	}
	data.CampaignEndTime.Value = endTime.String()
	data.CampaignEndTime.Nextin = int(endTime.Sub(time.Now()).Hours())
	data.Discount.Value = money["discount"][0]
	data.GmvActiveSession.CurrencyType = store.Currency
	data.GmvActiveSession.Value = money["gmv"][0]
	data.GmvActiveSession.GmvVertical.Positive = money["gmv"][0] >= money["gmv"][1]
	data.GmvActiveSession.GmvVertical.ChangePercentage = (float32(money["gmv"][0]-money["gmv"][1]) / float32(money["gmv"][1])) * 100
	data.ProductsActiveSession.Products = stats["products"][0]
	data.ProductsActiveSession.ProductsVertical.Positive = stats["products"][0] >= stats["products"][1]
	data.ProductsActiveSession.ProductsVertical.ChangePercentage = (float32(stats["products"][0]-stats["products"][1]) / float32(stats["products"][1])) * 100
	data.ActiveUsers.ActiveUsersInSession = stats["users"][0]
	data.ActiveUsers.ActiveUsersVertical.Positive = stats["users"][0] >= stats["users"][1]
	data.ActiveUsers.ActiveUsersVertical.ChangePercentage = (float32(stats["users"][0]-stats["users"][1]) / float32(stats["users"][1])) * 100

	// Marshal the map into a JSON string
	jsonData, err := json.Marshal(data)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

func (m *Repository) GetDealListActivity(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		DuratonType string `json:"durationType"`
	}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		m.App.ErrorLog.Println(err)
		return
	}

	duration := requestBody.DuratonType
	startTime := time.Now()
	endTime := time.Now()
	switch duration {
	case "12hours":
		startTime = startTime.Add(-12 * time.Hour)
		endTime = startTime.Add(-24 * time.Hour)
	case "24hours":
		startTime = startTime.Add(-24 * time.Hour)
		endTime = startTime.Add(-48 * time.Hour)
	case "weekly":
		startTime = startTime.Add(-24 * 7 * time.Hour)
		endTime = startTime.Add(-48 * 7 * time.Hour)
	case "monthly":
		startTime = startTime.Add(-24 * 30 * time.Hour)
		endTime = startTime.Add(-48 * 30 * time.Hour)
	}

	money, err := m.DB.GetDealDataFromCheckout(startTime, endTime, storeid)
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch deal data from checkout")
		helpers.ServerError(w, err)
	}
	data, err := m.DB.GetDealDataFromVisitor(startTime, endTime, storeid)
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch deal data from visitor")
		helpers.ServerError(w, err)
	}

	moneySeries, err := m.DB.GetSeriesDataFromCheckout(startTime, storeid)
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch series data from checkout")
		helpers.ServerError(w, err)
	}
	dataSeries, err := m.DB.GetSeriesDataFromVisitor(startTime, storeid)
	if err != nil {
		m.App.ErrorLog.Println("Failed to fetch series data from visitor")
		helpers.ServerError(w, err)
	}

	stats := models.DealListActivity{}

	stats.DiscountSpends.Price = money["discount"][0]
	stats.DiscountSpends.DiscountSpendsVertical.DiscountSpendsVertical = money["discount"][0] > money["discount"][1]
	stats.DiscountSpends.DiscountSpendsVertical.VerticalVal = (float32(money["discount"][0]-money["discount"][1]) / float32(money["discount"][1])) * 100
	stats.Gmv.Price = money["gmv"][0]
	stats.Gmv.GmvVertical.GmvVertical = money["gmv"][0] >= money["gmv"][1]
	stats.Gmv.GmvVertical.VerticalVal = (float32(money["gmv"][0]-money["gmv"][1]) / float32(money["gmv"][1])) * 100
	stats.Products.Price = data["products"][0]
	stats.Products.ProductsVertical.ProductsVertical = data["products"][0] >= data["products"][1]
	stats.Products.ProductsVertical.VerticalVal = (float32(data["products"][0]-data["products"][1]) / float32(data["products"][1])) * 100
	stats.Users.Price = data["users"][0]
	stats.Users.UsersVertical.UsersVertical = data["users"][0] >= data["users"][1]
	stats.Users.UsersVertical.VerticalVal = (float32(data["users"][0]-data["users"][1]) / float32(data["users"][1])) * 100

	stats.GmvData = moneySeries[0]
	stats.DiscountsData = moneySeries[1]
	stats.ProductsData = dataSeries[1]
	stats.UsersData = dataSeries[0]

	jsonData, err := json.Marshal(stats)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

func (m *Repository) GetTrendingProducts(w http.ResponseWriter, r *http.Request) {
	/* Initialize the function with the user and store context from the session */
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store

	/* DB API to fetch the most added products by end customers */
	list, deals, discounts, gmv, err := m.DB.GetTopProducts(storeid)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	data := models.TopProducts{}
	store, err := m.DB.GetStoreByID(storeid)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	for i, product := range list {
		/* Get product from shopify to fetch images etc
		We can pre store this info while initializing in the DB as well
		Since we get only 2 things from shopify: title and image
		*/
		p, err := store.GetProductById(product)
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
		prod.Users = deals[i]
		prod.Discount.CurrencyType = store.Currency
		prod.Discount.Value = discounts[i]
		prod.Gmv.CurrencyType = store.Currency
		prod.Gmv.Value = gmv[i]

		data.Products = append(data.Products, prod)
	}

	// Marshal the map into a JSON string
	jsonData, err := json.Marshal(data)
	if err != nil {
		m.App.ErrorLog.Println(err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

func (m *Repository) GetOtfVisitorData(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		DurationType string `json:"durationType"`
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

	data, err := m.DB.GetAggOtfByDuration(startTime, storeid)
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

func (m *Repository) GetAllCampaigns(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store

	campaigns, err := m.DB.GetAllCampaigns(storeid)
	if err != nil {
		m.App.ErrorLog.Println("Unable to fetch all campaigns")
		helpers.ServerError(w, err)
	}

	jsonData, err := json.Marshal(campaigns)
	if err != nil {
		m.App.ErrorLog.Println(err)
		helpers.ServerError(w, err)
	}

	fmt.Fprintf(w, "%s\n", jsonData)
}

func (m *Repository) ConfigureDiscounts(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		DefaultDiscount  string `json:"default_discount"`
		DiscountCateogry string `json:"discount_category"`
	}

	if err := json.Unmarshal(body, &requestBody); err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		m.App.ErrorLog.Println(err)
		return
	}
}

func (m *Repository) ConfigureDealList(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store
}

func (m *Repository) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store
}

func (m *Repository) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	user := m.App.Session.Get(r.Context(), "user").(models.User)
	storeid := user.Store
}

func (m *Repository) GetOtfUserInfo(w http.ResponseWriter, r *http.Request) {
	// we just pull the anonymousID and then pull the aggregate from click fucking house
	body, err := io.ReadAll(r.Body)
	if err != nil {
		m.App.ErrorLog.Panicln("Failed to read request body:", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	var requestBody struct {
		AnonymousID string `json:"anonymousid"`
	}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		m.App.ErrorLog.Panicln("Failed to Parse JSON")
	}
	vt, err := m.Clickhouse.PullStreamByAnonymousID(requestBody.AnonymousID)
	if err != nil {
		helpers.ServerError(w, err)
		m.App.ErrorLog.Panicln("Clickhouse pe data naas", err)
	}
	// once we have that, we cacluate otf and respond!
	otf, err := helpers.GetOtf(vt)
	if err != nil {
		helpers.ServerError(w, err)
		m.App.ErrorLog.Panicln("OTF naas", err)
	}
	if otf {
		fmt.Fprintln(w, "Dikhao BC!")
	} else {
		fmt.Fprintln(w, "Mat dikaho BC!")
	}
	// then we store in postgres --> this should take 2 secs max, usse zyada liya to ma chud jaegi

}

// if login window and func for login post method is handled
// do this first
// m.app.Session.RenewToken(r.Context())
// then put the user in the session
