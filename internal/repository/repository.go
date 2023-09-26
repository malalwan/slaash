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
	GetAllCampaigns(id int) ([]models.Camapign, error)
	UpdateDealListConfig(id int, md int8, pc string, bs int8, bc string) error
	GetStoreByID(id int) (models.Store, error)
	// CreateStore(s models.Store) error
	// UpdateStore(s models.Store) (models.Store, error)
}

type ClickhouseRepo interface {
	AllUsers()
	PullStreamByAnonymousID(id string) (models.VisitTable, error)
}
