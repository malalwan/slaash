package dbrepo

import (
	"context"
	"time"

	"github.com/malalwan/slaash/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

/* Function to retrieve all products that are part of a specific campaign */
func (m *postgresDBRepo) GetCampaignProducts(c models.Campaign) ([]models.CampaignProduct, error) {
	// pull all campaign products with the given campaing ID from DB
	/* SELECT *
	FROM campaign_product
	WHERE CampaignID = 1;
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `select * from campaign_product where 
			campaignid = $1`

	rows, err := m.DB.QueryContext(ctx, stmt, c.CampaignID)
	if err != nil {
		return []models.CampaignProduct{}, err
	}
	defer rows.Close()

	var campaignProducts []models.CampaignProduct

	for rows.Next() {
		var cp models.CampaignProduct
		err := rows.Scan(&cp.ID, &cp.CampaignID, &cp.ProductID, &cp.Title,
			&cp.Store.ID, &cp.Deals, &cp.Sold, &cp.DealDiscount,
			&cp.EmailSentTo, &cp.Misc, &cp.PriceRuleID)
		if err != nil {
			return []models.CampaignProduct{}, err
		}
		campaignProducts = append(campaignProducts, cp)
	}
	if err := rows.Err(); err != nil {
		return []models.CampaignProduct{}, err
	}

	return campaignProducts, nil
}

func (m *postgresDBRepo) GetActiveCampaign(s models.Store) ([]models.Campaign, error) {
	// pull active campaign(s) for a store from DB
	/* SELECT *
	FROM campaign
	WHERE storeid = 1;
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `select * from campaign where 
			storeid = $1`

	rows, err := m.DB.QueryContext(ctx, stmt, s.ID)
	if err != nil {
		return []models.Campaign{}, err
	}
	defer rows.Close()

	var campaigns []models.Campaign

	for rows.Next() {
		var c models.Campaign
		err := rows.Scan(&c.CampaignID, &c.Store.ID, &c.Timestamp,
			&c.Discount, &c.ActiveStatus, &c.Misc)
		if err != nil {
			return []models.Campaign{}, err
		}
		campaigns = append(campaigns, c)
	}
	if err := rows.Err(); err != nil {
		return []models.Campaign{}, err
	}

	return campaigns, nil
}

func (m *postgresDBRepo) CreateCampaign(c models.Campaign) error {
	// create a campign and return it
	/* INSERT INTO campaign (storeid, timestamp, discount, activestatus, misc)
	VALUES ($1, $2, $3, $4, $5)
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `select * from campaign where 
			storeid = $1`

	_, err := m.DB.ExecContext(ctx, stmt, c.Store.ID, time.Now(), c.Discount, c.ActiveStatus, c.Misc)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) CreateCampaignProducts(c models.Campaign) ([]models.CampaignProduct, error) {
	// create a campign and return it
	/* INSERT INTO campaign (storeid, timestamp, discount, activestatus, misc)
	VALUES ($1, $2, $3, $4, $5)
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `select * from campaign where
			storeid = $1`

	_, err := m.DB.ExecContext(ctx, stmt, c.Store.ID, time.Now(), c.Discount, c.ActiveStatus, c.Misc)
	if err != nil {
		return []models.CampaignProduct{}, err
	}

	return []models.CampaignProduct{}, nil
}

func (m *postgresDBRepo) GetCampaignByID(int64) (models.Campaign, error) {
	return models.Campaign{}, nil
}

func (m *postgresDBRepo) UpdateCampaignProducts(models.Campaign) ([]models.CampaignProduct, error) {
	return []models.CampaignProduct{}, nil
}

func (m *postgresDBRepo) ListAllCampaigns() ([]models.Campaign, error) {
	return []models.Campaign{}, nil
}

func (m *postgresDBRepo) GetStoreByID() ([]models.Campaign, error) {
	return []models.Campaign{}, nil
}
