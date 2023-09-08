package models

import (
	"fmt"
	"log"

	goshopify "github.com/bold-commerce/go-shopify/v3"
	"github.com/malalwan/slaash/internal/config"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewShopifyFunctions(a *config.AppConfig) {
	app = a
}

func (store Store) SendUItoTheme(js string) error {
	client := store.InitClient()
	themes, err := client.Theme.List(client.Theme)
	if err != nil {
		fmt.Println("Error retrieving themes:", err)
		return err
	}

	var activeThemeID int64
	for _, theme := range themes {
		if theme.Role == "main" { // 'main' role indicates the active theme
			activeThemeID = theme.ID
			break
		}
	}

	if activeThemeID == 0 {
		fmt.Println("Active theme not found")
		return err
	}

	_, err = client.Theme.Get(activeThemeID, client.Theme)
	if err != nil {
		fmt.Println("Error retrieving active theme:", err)
		return err
	}

	assetKey := "assets/global.js"

	asset := goshopify.Asset{
		ThemeID: activeThemeID,
		Value:   js,
		Key:     assetKey,
	}

	_, err = client.Asset.Update(activeThemeID, asset)

	if err != nil {
		fmt.Println("Error here!")
		log.Fatal(err)
		return err
	}

	fmt.Println("Updated asset name!")

	return nil
}

func (store Store) CreatePriceRule(pr goshopify.PriceRule) (int64, error) {

	client := store.InitClient()

	newPriceRule, err := goshopify.PriceRuleService.Create(client.PriceRule, pr)
	if err != nil {
		log.Fatalf("Failed to create price rule: %s", err)
	}

	fmt.Printf("Price rule created with ID: %d\n", newPriceRule.ID)
	// DB calls here

	return newPriceRule.ID, nil
}

func (store Store) SendJsToGlobal(js string) error {
	client := store.InitClient()
	themes, err := client.Theme.List(client.Theme)
	if err != nil {
		fmt.Println("Error retrieving themes:", err)
		return err
	}

	var activeThemeID int64
	for _, theme := range themes {
		if theme.Role == "main" { // 'main' role indicates the active theme
			activeThemeID = theme.ID
			break
		}
	}

	if activeThemeID == 0 {
		fmt.Println("Active theme not found")
		return err
	}

	_, err = client.Theme.Get(activeThemeID, client.Theme)
	if err != nil {
		fmt.Println("Error retrieving active theme:", err)
		return err
	}

	assetKey := "assets/global-slaash.js"

	asset := goshopify.Asset{
		ThemeID: activeThemeID,
		Value:   js,
		Key:     assetKey,
	}

	_, err = client.Asset.Update(activeThemeID, asset)

	if err != nil {
		fmt.Println("Error here!")
		return err
	}

	fmt.Println("Updated asset name!")

	return nil
}

func (store Store) FetchPriceRules() ([]goshopify.PriceRule, error) {

	client := store.InitClient()

	priceRuleList, err := goshopify.PriceRuleService.List(client.PriceRule)

	if err != nil {
		log.Fatalf("Failed to fetch price rules: %s", err)
	}

	return priceRuleList, nil
}

func (store Store) CreateDiscountByPrID(prId int64, d goshopify.PriceRuleDiscountCode) (int, error) {

	client := store.InitClient()

	newD, err := goshopify.DiscountCodeService.Create(client.DiscountCode, prId, d)
	if err != nil {
		log.Fatalf("Failed to create price rule: %s", err)
	}

	fmt.Printf("Discount Code created with ID: %d\n", newD.ID)
	// DB calls here

	return 0, nil
}

func (store Store) DeleteDiscountByDiscId(dId int64, prId int64) error {

	client := store.InitClient()

	err := goshopify.DiscountCodeService.Delete(client.DiscountCode, dId, prId)
	if err != nil {
		log.Fatalf("Failed to delete discount code: %s", err)
	}
	return nil
}

func (store Store) FetchDiscountsByPrId(prId int64) ([]goshopify.PriceRuleDiscountCode, error) {

	client := store.InitClient()

	dList, err := goshopify.DiscountCodeService.List(client.DiscountCode, prId)
	if err != nil {
		log.Fatalf("Failed to create price rule: %s", err)
	}
	return dList, nil
}

func (store Store) GetOrderData() ([]goshopify.Order, error) {

	client := store.InitClient()
	var intf interface{}
	orders, err := goshopify.OrderService.List(client.Order, intf)
	if err != nil {
		log.Fatalf("Failed to retrieve order list: %s", err)
	}
	return orders, nil
}

func (store Store) GetCustomerByCustId(CustId int64) (*goshopify.Customer, error) {

	client := store.InitClient()

	customer, err := goshopify.CustomerService.Get(client.Customer, CustId, 0)

	return customer, err
}

func (store Store) RetrieveAbandonedCheckouts() ([]goshopify.AbandonedCheckout, error) {

	client := store.InitClient()

	AbanCheckouts, err := goshopify.AbandonedCheckoutService.List(client.AbandonedCheckout, 0)

	return AbanCheckouts, err
}

func (store Store) GetProductById(PId int64) (*goshopify.Product, error) {

	client := store.InitClient()

	product, err := goshopify.ProductService.Get(client.Product, PId, 0)

	return product, err
}

func (store Store) InitClient() *goshopify.Client {
	app := goshopify.App{
		ApiKey:    app.MyAppCreds[0],
		ApiSecret: app.MyAppCreds[1],
	}
	client := goshopify.NewClient(app, store.Name, store.ApiToken)
	return client
}

func (store Store) GetAllProducts() ([]goshopify.Product, error) {

	client := store.InitClient()

	products, err := goshopify.ProductService.List(client.Product, 0)

	return products, err

}

func (store Store) GetAllCustomers() ([]goshopify.Customer, error) {

	client := store.InitClient()

	customers, err := goshopify.CustomerService.List(client.Customer, nil)

	if err != nil {
		log.Fatalf("Failed to retrieve order list: %s", err)
	}
	return customers, nil
}

func (store Store) GetOrdersByCustomerId(CustId int64) ([]goshopify.Order, error) {

	client := store.InitClient()

	orders, err := goshopify.CustomerService.ListOrders(client.Customer, CustId, 0)

	// goshopify.Order
	return orders, err
}

func (store Store) CreateWebhook(w goshopify.Webhook) (*goshopify.Webhook, error) {
	client := store.InitClient()

	// sample webhook:
	// webhook := shopify.Webhook{
	// 	Topic:   "products/create",
	// 	Address: "https://your-webhook-endpoint.com/webhook",
	// 	Format:  "json",
	// }

	newW, err := goshopify.WebhookService.Create(client.Webhook, w)

	return newW, err
}

// Webhook to get a notification when a checkout happens or is dropped

func (store Store) RetrieveAllWebhooks() ([]goshopify.Webhook, error) {
	client := store.InitClient()

	listOptions := goshopify.ListOptions{
		Limit: 10, // Number of items per page
		Order: "desc",
		// Add other options if needed
	}

	webhooks, error := goshopify.WebhookService.List(client.Webhook, listOptions)

	return webhooks, error
}
