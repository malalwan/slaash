package dbrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/malalwan/slaash/internal/models"
)

/* Table "public.campaign"
    Column    |  Type
--------------+--------
 campaignid   | integer
 storeid      | integer
 timestamp    | timestamp without time zone
 discount     | integer
 activestatus | integer
 misc         | text
*/

func (m *postgresDBRepo) GetActiveCampaign(s models.Store) (models.Campaign, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `select * from campaign where 
			storeid = $1`

	rows, err := m.DB.QueryContext(ctx, stmt, s.ID)
	if err != nil {
		return models.Campaign{}, err
	}
	defer rows.Close()

	var campaigns []models.Campaign

	for rows.Next() {
		var c models.Campaign
		err := rows.Scan(&c.CampaignID, &c.Store.ID, &c.Timestamp,
			&c.Discount, &c.ActiveStatus, &c.Misc)
		if err != nil {
			return models.Campaign{}, err
		}
		campaigns = append(campaigns, c)
	}
	if err := rows.Err(); err != nil {
		return models.Campaign{}, err
	}

	return campaigns[0], nil
}

func (m *postgresDBRepo) CreateCampaign(c models.Campaign) error {
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

func (m *postgresDBRepo) SelectFromCampaignById(id int64, ts time.Time, s string, f string, w string) (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := fmt.Sprintf("select %s from %s where %s", s, f, w)
	j := make(map[string]int)
	rows, err := m.DB.QueryContext(ctx, stmt, id, ts)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		var g, p, u, d sql.NullInt64
		err = rows.Scan(&g, &p, &u, &d)
		if !g.Valid {
			j["gmv"] = 0
			return j, nil
		}
		if err != nil {
			return j, err
		}
		j["gmv"] = int(g.Int64)
		j["products"] = int(p.Int64)
		j["users"] = int(u.Int64)
		j["discounts"] = int(d.Int64)
	}

	return j, nil
}

func (m *postgresDBRepo) GetGroupSeriesData(id int64, ts time.Time) ([]map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	gmap := make(map[string]int)
	pmap := make(map[string]int)
	umap := make(map[string]int)
	dmap := make(map[string]int)
	stmt := `SELECT DATE_TRUNC('hour', timestamp) AS interval,
	COALESCE(SUM(price),0), COALESCE(COUNT(*), 0) ,
	COALESCE(SUM(deals),0), COALESCE(SUM(dealdiscount*price/100),0)
	FROM   campaign_product
	WHERE  storeid = $1
	AND	  timestamp >= $2
	GROUP BY interval
	ORDER BY interval;`
	rows, err := m.DB.QueryContext(ctx, stmt, id, ts)
	if err != nil {
		return []map[string]int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var t time.Time
		var g, p, u, d sql.NullInt64
		err = rows.Scan(&t, &g, &p, &u, &d)
		if err != nil {
			return []map[string]int{}, err
		}
		gmap[t.Format("2006-01-02 15:04:05")] = int(g.Int64)
		pmap[t.Format("2006-01-02 15:04:05")] = int(p.Int64)
		umap[t.Format("2006-01-02 15:04:05")] = int(u.Int64)
		dmap[t.Format("2006-01-02 15:04:05")] = int(d.Int64)
	}

	return []map[string]int{gmap, pmap, umap, dmap}, nil
}

func (m *postgresDBRepo) GetCampaignByID(id int64) (models.Campaign, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "select * from campaign where campaignid = $1"
	j := models.Campaign{}
	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&j.CampaignID, &j.Store.ID, &j.Timestamp, &j.Discount, &j.ActiveStatus, &j.Misc)
		if err != nil {
			return j, err
		}
	}
	return j, nil
}

func (m *postgresDBRepo) ListAllCampaigns(s int) ([]models.Campaign, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "select * from campaign where storeid = $1"
	j := []models.Campaign{}
	rows, err := m.DB.QueryContext(ctx, stmt, s)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Campaign
		err = rows.Scan(&c.CampaignID, &c.Store.ID, &c.Timestamp, &c.Discount, &c.ActiveStatus, &c.Misc)
		if err != nil {
			return j, err
		}
		st, err := m.GetStoreByID(c.Store.ID)
		if err != nil {
			return j, err
		}
		c.Store = st
		j = append(j, c)
	}

	return j, nil
}

/* Table "public.users"
	Column   |   Type
-------------+-----------
 id          | integer
 firstname   | text
 lastname    | text
 email       | text
 password    | text
 accesslevel | integer
 createdat   | timestamp without time zone
 updatedat   | timestamp without time zone
 storeid     | integer
*/

func (m *postgresDBRepo) CreateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into users 
				(firstname, lastname, email, password,
				accesslevel, createdat, updatedat, storeid)
			 values
			 	($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := m.DB.ExecContext(ctx, stmt, u.FirstName, u.LastName, u.Email,
		u.Password, u.AccessLevel, u.CreatedAt,
		u.UpdatedAt, u.Store.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) UpdateUser(u models.User) (models.User, error) {
	return models.User{}, nil
}

func (m *postgresDBRepo) GetUserByStore(storeid int) (models.User, error) {
	return models.User{}, nil
}

/* Table "public.price_rule"
      Column       |  Type
-------------------+--------
 id                | integer
 targettype        | text
 targetselection   | text
 valuetype         | text
 value             | numeric
 customerselection | text
 allocationmethod  | text
 startsat          | timestamp without time zone
*/

func (m *postgresDBRepo) CreatePr(pr models.PriceRule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into price_rule 
				(prid, targettype, targetselection, valuetype,
				value, customerselection, allocationmethod, startsat)
			 values
			 	($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := m.DB.ExecContext(ctx, stmt, pr.ID, pr.TargetType, pr.TargetSelection,
		pr.ValueType, pr.Value, pr.CustomerSelection, pr.AllocationMethod, pr.StartsAt)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) DeletePr(pr models.PriceRule) error {
	return nil
}
func (m *postgresDBRepo) GetPrById(id int64, storeid int) (models.PriceRule, error) {
	return models.PriceRule{}, nil
}
func (m *postgresDBRepo) ListPr(storeid int) ([]models.PriceRule, error) {
	return []models.PriceRule{}, nil
}

/* Table "public.discount_code"
	Column   |   Type
 ------------+-----------
 id          | integer
 priceruleid | integer
 code        | text
 usagecount  | integer
 createdat   | timestamp without time zone
 updatedat   | timestamp without time zone
*/

func (m *postgresDBRepo) CreateDiscountCode(d models.DiscountCode) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into discount_code
				(priceruleId, code, usagecount,
				createdat, updatedat)
			 values
			 	($1, $2, $3, $4, $5)`

	_, err := m.DB.ExecContext(ctx, stmt, d.PriceRuleID, d.Code,
		d.UsageCount, d.Timestamp, d.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) DeleteDiscountCode(d models.DiscountCode) error {
	return nil
}

func (m *postgresDBRepo) GetDiscountsByPr(pr models.PriceRule) ([]models.DiscountCode, error) {
	return []models.DiscountCode{}, nil
}

/* Table "public.buyer"
   Column    |  Type
-------------+--------
 anonymousid | integer
 email       | text
 storeid     | integer
 productid   | integer
 timestamp   | timestamp without time zone
 gotdeal     | boolean
 clickeddeal | boolean
 cpid        | integer
 misc        | text
*/

func (m *postgresDBRepo) CreateBuyer(b models.Buyer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into buyer 
				(anonymousid, email, storeid, productid,
					timestamp, gotdeal, clickeddeal, cpid, misc)
			 values
			 	($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := m.DB.ExecContext(ctx, stmt, b.AnonymousID, b.Email,
		b.Store.ID, b.ProductId, b.Timestamp, b.GotDeal, b.ClickedDeal,
		b.CPID, b.Misc)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) GetBuyersByStore(storeid int) ([]models.Buyer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "select * from buyer where storeid = $1"
	j := []models.Buyer{}
	rows, err := m.DB.QueryContext(ctx, stmt, storeid)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		var b models.Buyer
		err = rows.Scan(&b.AnonymousID, &b.Email, &b.Store.ID, &b.ProductId,
			&b.Timestamp, &b.GotDeal, &b.ClickedDeal, &b.CPID, &b.Misc)
		if err != nil {
			return j, err
		}
		j = append(j, b)
	}
	return j, nil
}

func (m *postgresDBRepo) UpdateBuyer(b models.Buyer) (models.Buyer, error) {
	return models.Buyer{}, nil
}

func (m *postgresDBRepo) GetAggregateOtfByDuration(ts time.Time, typ string, id int) (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	otfMap := map[string]int{}

	stmt := `SELECT DATE_TRUNC('hour', timestamp) AS interval,
	COALESCE(COUNT(*), 0)
	FROM   buyer
	WHERE  storeid = $1
	AND	  timestamp >= $2
	GROUP BY interval
	ORDER BY interval;`
	rows, err := m.DB.QueryContext(ctx, stmt, id, ts)
	if err != nil {
		return otfMap, err
	}
	defer rows.Close()
	for rows.Next() {
		var t time.Time
		var otf sql.NullInt64
		err = rows.Scan(&t, &otf)
		if err != nil {
			return otfMap, err
		}
		otfMap[t.Format("2006-01-02 15:04:05")] = int(otf.Int64)
	}

	return otfMap, nil
}

/* Table "public.store"
    Column    |  Type
--------------+--------
 id           | integer
 name         | text
 apitoken     | text
 refreshtoken | text
 misc         | text
 url          | text
*/

func (m *postgresDBRepo) CreateStore(s models.Store) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into store 
				(name, apitoken, refreshtoken,
				misc, url)
			 values
			 	($1, $2, $3, $4, $5)`

	_, err := m.DB.ExecContext(ctx, stmt, s.Name, s.ApiToken,
		s.RefreshToken, s.Misc, s.URL)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) GetStoreByID(id int) (models.Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "select * from store where id = $1"
	j := models.Store{}
	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&j.ID, &j.Name, &j.ApiToken, &j.RefreshToken, &j.Misc, &j.URL)
		if err != nil {
			return j, err
		}
	}
	return j, nil
}

func (m *postgresDBRepo) UpdateStore(s models.Store) (models.Store, error) {
	return models.Store{}, nil
}

/* Table "public.campaign_product"
    Column    |  Type
--------------+--------
 id           | integer
 campaignid   | integer
 productid    | integer
 title        | text
 storeid      | integer
 deals        | integer
 sold         | integer
 dealdiscount | integer
 emailsentto  | text[]
 misc         | text
 priceruleid  | integer
 timestamp    | timestamp without time zone
 price        | integer
*/

func (m *postgresDBRepo) CreateCampaignProduct(cp models.CampaignProduct) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into campaign_product 
				(campaignid, productid, title, storeid,
				deals, sold, dealdiscount, misc, 
				timestamp, price)
			 values
			 	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := m.DB.ExecContext(ctx, stmt, cp.CampaignID, cp.ProductID,
		cp.Title, cp.Store.ID, cp.Deals, cp.Sold, cp.DealDiscount,
		cp.Misc, cp.Timestamp, cp.Price)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) GetCampaignProducts(c int64) ([]models.CampaignProduct, error) {
	// pull all campaign products with the given campaing ID from DB
	/* SELECT *
	FROM campaign_product
	WHERE CampaignID = 1;
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `select * from campaign_product where 
			campaignid = $1`

	rows, err := m.DB.QueryContext(ctx, stmt, c)
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
			return campaignProducts, err
		}
		campaignProducts = append(campaignProducts, cp)
	}
	if err := rows.Err(); err != nil {
		return campaignProducts, err
	}

	return campaignProducts, nil
}

func (m *postgresDBRepo) UpdateCampaignProducts(c models.Campaign, dict map[string]interface{}) ([]models.CampaignProduct, error) {
	return []models.CampaignProduct{}, nil
}

func (m *postgresDBRepo) GetTopProductsByStore(s int) ([]models.CampaignProduct, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	/* SELECT * FROM products
	ORDER BY price ASC
	LIMIT 5;
	*/
	products := []models.CampaignProduct{}
	stmt := `select productid, deals, dealdiscount, price 
	from campaign_product where storeid = $1 order by deals desc limit 5`
	rows, err := m.DB.QueryContext(ctx, stmt, s)
	if err != nil {
		return products, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid, d, dd, p sql.NullInt64

		err = rows.Scan(&pid, &d, &dd, &p)
		if !pid.Valid || err != nil {
			return products, err
		}
		var c models.CampaignProduct
		c.ProductID = pid.Int64
		c.Deals = int(d.Int64)
		c.DealDiscount = int(dd.Int64)
		c.Price = int(p.Int64)
		products = append(products, c)
	}

	return products, nil
}
