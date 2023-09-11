package repository

import (
	"time"

	"github.com/malalwan/slaash/internal/models"
)

type DatabaseRepo interface {
	// campaign_product
	CreateCampaignProduct(cp models.Campaign) error
	GetCampaignProducts(c int64) ([]models.CampaignProduct, error)
	UpdateCampaignProducts(c models.Campaign, dict map[string]interface{}) ([]models.CampaignProduct, error)
	// camapign
	GetActiveCampaign(s models.Store) ([]models.Campaign, error)
	CreateCampaign(c models.Campaign) error
	GetCampaignByID(id int64) (models.Campaign, error)
	ListAllCampaigns(storeid int) ([]models.Campaign, error)
	SelectFromCampaignByStore(storeid int, ts time.Time, s string, f string, w string) (map[string]int, error)
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
