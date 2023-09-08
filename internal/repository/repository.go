package repository

import (
	"github.com/malalwan/slaash/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool // Stub only
	GetCampaignProducts(c models.Campaign) ([]models.CampaignProduct, error)
	GetActiveCampaign(s models.Store) ([]models.Campaign, error)
	CreateCampaign(c models.Campaign) error
	CreateCampaignProducts(c models.Campaign) ([]models.CampaignProduct, error)
	GetCampaignByID(int64) (models.Campaign, error)
	UpdateCampaignProducts(models.Campaign) ([]models.CampaignProduct, error)
	ListAllCampaigns() ([]models.Campaign, error)
	GetStoreByID() ([]models.Campaign, error)
}
