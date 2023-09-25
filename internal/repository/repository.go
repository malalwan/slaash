package repository

import (
	"time"

	"github.com/malalwan/slaash/internal/models"
)

type DatabaseRepo interface {
	// new shit
	StopDealList(id int) error
	SetTurnOffTime(id int) error
	GetCampignEndTime(id int) (time.Time, error)
	GetAggFromCheckout(id int) (map[string][]int, error)
	GetAggFromVisitor(id int) (map[string][]int, error)
	GetDealDataFromCheckout(t1 time.Time, t2 time.Time, id int) (map[string][]int, error)
	GetDealDataFromVisitor(t1 time.Time, t2 time.Time, id int) (map[string][]int, error)
	GetSeriesDataFromCheckout(t time.Time, id int) ([]map[string]int, error)
	GetSeriesDataFromVisitor(t time.Time, id int) ([]map[string]int, error)
	GetTopProducts(id int) ([]int64, []int, []int, []int, error)
	GetAggOtfByDuration(ts time.Time, id int) (map[string]int, error)
	GetAllCampaigns(storeid int) ([]models.Camapign, error)
	// campaign_product
	CreateCampaignProduct(cp models.CampaignProduct) error
	GetCampaignProducts(c int64) ([]models.CampaignProduct, error)
	UpdateCampaignProducts(c models.Campaign, dict map[string]interface{}) ([]models.CampaignProduct, error)
	GetTopProductsByStore(s int) ([]models.CampaignProduct, error)
	// camapign
	GetActiveCampaign(s models.Store) (models.Campaign, error)
	CreateCampaign(c models.Campaign) error
	GetCampaignByID(id int64) (models.Campaign, error)

	SelectFromCampaignById(id int64, ts time.Time, s string, f string, w string) (map[string]int, error)
	GetGroupSeriesData(id int64, ts time.Time) ([]map[string]int, error)
	// store
	GetStoreByID(id int) (models.Store, error)
	CreateStore(s models.Store) error
	UpdateStore(s models.Store) (models.Store, error)
	// buyer
	CreateBuyer(b models.Buyer) error
	UpdateBuyer(b models.Buyer) (models.Buyer, error)
	GetBuyersByStore(storeid int) ([]models.Buyer, error)
	// user
	CreateUser(u models.User) error
	UpdateUser(u models.User) (models.User, error)
	GetUserByStore(storeid int) (models.User, error)
	// price_rule
	CreatePr(pr models.PriceRule) error
	DeletePr(pr models.PriceRule) error
	GetPrById(id int64, storeid int) (models.PriceRule, error)
	ListPr(storeid int) ([]models.PriceRule, error)
	// discount_code
	CreateDiscountCode(d models.DiscountCode) error
	DeleteDiscountCode(d models.DiscountCode) error
	GetDiscountsByPr(pr models.PriceRule) ([]models.DiscountCode, error)
}

type ClickhouseRepo interface {
	AllUsers()
	PullStreamByAnonymousID(id string) (models.VisitTable, error)
}
