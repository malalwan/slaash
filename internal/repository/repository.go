package repository

import (
	"time"

	"github.com/malalwan/slaash/internal/models"
)

type DatabaseRepo interface {
	ToggleDealList(id int, t bool) error
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
	GetAllCampaigns(id int) ([]models.Campaign, error)
	GetStoreByID(id int) (models.Store, error)
	GetDefaultDiscountAndCategory(id int) (int8, int8, error)
	GetConfiguredDiscounts(id int, cat int8) (map[int64]int8, error)
	GetDealListInfo(id int) (models.DlInfo, error)
	GetUserProfileInfo(id int) (models.UserProfile, error)
	UpdateDiscounts(id int, dc int8, mp map[int64]int8) error
	UpdateDiscountDefaults(id int, def int8, cat int8) error
	UpdateDealListConfig(id int, md int8, pc string, bs int8, bc string) error
	UpdateUserProfile(id int, fn string, ln string, p string) error
	// CreateStore(s models.Store) error
	// UpdateStore(s models.Store) (models.Store, error)
}

type ClickhouseRepo interface {
	AllUsers()
	PullStreamByAnonymousID(id string) (models.VisitTable, error)
}
