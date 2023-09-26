package dbrepo

import (
	"context"
	"database/sql"
	"strconv"
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

func (m *postgresDBRepo) StopDealList(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE store
	 	 	 SET deal_list_active = FALSE
	 	 	 WHERE id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		m.App.ErrorLog.Println("DB insertion failed")
		return err
	}
	return nil
}

func (m *postgresDBRepo) SetTurnOffTime(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE store
			 SET campaign_turn_off_time = $1
			 WHERE id = $2`

	m.App.InfoLog.Println(time.Now().Format("15:04:05"))

	_, err := m.DB.ExecContext(ctx, stmt, time.Now().Format("15:04:05"), id)
	if err != nil {
		m.App.ErrorLog.Println("DB insertion failed")
		return err
	}
	return nil
}

func (m *postgresDBRepo) GetCampignEndTime(id int) (time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT campaign_renewal_time
			 FROM store
			 WHERE id = $1`

	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		m.App.ErrorLog.Println("DB extraction failed")
		return time.Now(), err
	}
	defer rows.Close()
	var s string
	var t time.Time
	for rows.Next() {
		err := rows.Scan(&s)
		if err != nil {
			m.App.ErrorLog.Panicln("Assignment of extracted value from DB failed")
			return time.Now(), err
		}

		h, _ := strconv.Atoi(s[:2])
		m, _ := strconv.Atoi(s[3:5])
		s, _ := strconv.Atoi(s[6:])
		ct := time.Now()
		ch, cm, cs := ct.Clock()
		after := false
		if h > ch {
			after = true
		} else if h == ch {
			if m > cm {
				after = true
			} else if m == cm {
				if s > cs {
					after = true
				} else if s == cs {
					after = true
				}
			}
		}

		if after {
			t = time.Date(ct.Year(), ct.Month(), ct.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		} else {
			t = time.Date(ct.Year(), ct.Month(), ct.Day()+1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
		}
	}

	// push t to db again
	return t, nil
}

func (m *postgresDBRepo) GetAggFromCheckout(id int) (map[string][]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mp := make(map[string][]int)

	stmt1 := `SELECT SUM(gmv), SUM(discount_amount)
			  FROM checkout
			  WHERE timestamp > $1 AND store = $2`

	stmt2 := `SELECT SUM(gmv), SUM(discount_amount)
			  FROM checkout
			  WHERE timestamp < $1 AND timestamp > $2 AND store = $3`

	t, err := m.GetCampignEndTime(id)
	if err != nil {
		return mp, err
	}

	rows1, err := m.DB.QueryContext(ctx, stmt1, t, id)
	if err != nil {
		return mp, err
	}
	defer rows1.Close()

	var gmv, disc sql.NullInt64
	for rows1.Next() {
		err = rows1.Scan(&gmv, &disc)
		if err != nil {
			return mp, err
		}
	}

	mp["gmv"] = []int{}
	mp["discount"] = []int{}

	if gmv.Valid && disc.Valid {
		mp["gmv"] = append(mp["gmv"], int(gmv.Int64))
		mp["discount"] = append(mp["discount"], int(disc.Int64))
	} else {
		mp["gmv"] = append(mp["gmv"], 0)
		mp["discount"] = append(mp["discount"], 0)
	}

	rows2, err := m.DB.QueryContext(ctx, stmt2, t, t.AddDate(0, 0, -1), id)
	if err != nil {
		return mp, err
	}
	defer rows2.Close()

	for rows1.Next() {
		err = rows1.Scan(&gmv, &disc)
		if err != nil {
			return mp, err
		}
	}

	if gmv.Valid && disc.Valid {
		mp["gmv"] = append(mp["gmv"], int(gmv.Int64))
		mp["discount"] = append(mp["discount"], int(disc.Int64))
	} else {
		mp["gmv"] = append(mp["gmv"], 0)
		mp["discount"] = append(mp["discount"], 0)
	}

	return mp, nil
}

func (m *postgresDBRepo) GetAggFromVisitor(id int) (map[string][]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mp := make(map[string][]int)

	stmt1 := `SELECT COUNT(anonymous_id), COUNT(DISTINCT product_id)
			  FROM visitor
			  WHERE timestamp > $1 AND store = $2`

	stmt2 := `SELECT COUNT(anonymous_id), COUNT(DISTINCT product_id)
			  FROM visitor
			  WHERE timestamp < $1 AND timestamp > $2 AND store = $3`

	t, err := m.GetCampignEndTime(id)
	if err != nil {
		return mp, err
	}

	rows1, err := m.DB.QueryContext(ctx, stmt1, t, id)
	if err != nil {
		return mp, err
	}
	defer rows1.Close()

	var users, products int
	for rows1.Next() {
		err = rows1.Scan(&users, &products)
		if err != nil {
			return mp, err
		}
	}

	mp["users"] = []int{}
	mp["products"] = []int{}
	mp["users"] = append(mp["users"], users)
	mp["products"] = append(mp["products"], products)

	rows2, err := m.DB.QueryContext(ctx, stmt2, t, t.AddDate(0, 0, -1), id)
	if err != nil {
		return mp, err
	}
	defer rows2.Close()

	for rows1.Next() {
		err = rows1.Scan(&users, &products)
		if err != nil {
			return mp, err
		}
	}

	mp["users"] = append(mp["users"], users)
	mp["products"] = append(mp["products"], products)

	return mp, nil
}

func (m *postgresDBRepo) GetDealDataFromCheckout(t1 time.Time, t2 time.Time, id int) (map[string][]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mp := make(map[string][]int)

	stmt1 := `SELECT SUM(gmv), SUM(discount_amount)
			  FROM checkout
			  WHERE timestamp > $1 AND store = $2`

	stmt2 := `SELECT COUNT(anonymous_id), COUNT(DISTINCT product_id)
			  FROM visitor
			  WHERE timestamp < $1 AND timestamp > $2 AND store = $3`

	rows1, err := m.DB.QueryContext(ctx, stmt1, t1, id)
	if err != nil {
		return mp, err
	}
	defer rows1.Close()

	var gmv, disc sql.NullInt64
	for rows1.Next() {
		err = rows1.Scan(&gmv, &disc)
		if err != nil {
			return mp, err
		}
	}

	mp["gmv"] = []int{}
	mp["discount"] = []int{}
	if gmv.Valid && disc.Valid {
		mp["gmv"] = append(mp["gmv"], int(gmv.Int64))
		mp["discount"] = append(mp["discount"], int(disc.Int64))
	} else {
		mp["gmv"] = append(mp["gmv"], 0)
		mp["discount"] = append(mp["discount"], 0)
	}

	rows2, err := m.DB.QueryContext(ctx, stmt2, t1, t2, id)
	if err != nil {
		return mp, err
	}
	defer rows2.Close()

	for rows1.Next() {
		err = rows1.Scan(&gmv, &disc)
		if err != nil {
			return mp, err
		}
	}

	if gmv.Valid && disc.Valid {
		mp["gmv"] = append(mp["gmv"], int(gmv.Int64))
		mp["discount"] = append(mp["discount"], int(disc.Int64))
	} else {
		mp["gmv"] = append(mp["gmv"], 0)
		mp["discount"] = append(mp["discount"], 0)
	}

	return mp, nil
}

func (m *postgresDBRepo) GetDealDataFromVisitor(t1 time.Time, t2 time.Time, id int) (map[string][]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mp := make(map[string][]int)

	stmt1 := `SELECT COUNT(anonymous_id), COUNT(DISTINCT product_id)
			  FROM visitor
			  WHERE timestamp > $1 AND store = $2`

	stmt2 := `SELECT COUNT(anonymous_id), COUNT(DISTINCT product_id)
			  FROM visitor
			  WHERE timestamp < $1 AND timestamp > $2 AND store = $3`

	rows1, err := m.DB.QueryContext(ctx, stmt1, t1, id)
	if err != nil {
		return mp, err
	}
	defer rows1.Close()

	var users, products int
	for rows1.Next() {
		err = rows1.Scan(&users, &products)
		if err != nil {
			return mp, err
		}
	}

	mp["users"] = []int{}
	mp["products"] = []int{}
	mp["users"] = append(mp["users"], users)
	mp["products"] = append(mp["products"], products)

	rows2, err := m.DB.QueryContext(ctx, stmt2, t1, t2, id)
	if err != nil {
		return mp, err
	}
	defer rows2.Close()

	for rows1.Next() {
		err = rows1.Scan(&users, &products)
		if err != nil {
			return mp, err
		}
	}

	mp["users"] = append(mp["users"], users)
	mp["products"] = append(mp["products"], products)

	return mp, nil
}

func (m *postgresDBRepo) GetSeriesDataFromCheckout(t time.Time, id int) ([]map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	gmap := make(map[string]int)
	dmap := make(map[string]int)

	stmt := `SELECT DATE_TRUNC('hour', timestamp) AS interval,
			 COALESCE(SUM(gmv),0), COALESCE(SUM(discount_amount),0)
			 FROM checkout
		     WHERE store = $1
			 AND timestamp >= $2
			 GROUP BY interval
			 ORDER BY interval;`

	rows, err := m.DB.QueryContext(ctx, stmt, id, t)
	if err != nil {
		return []map[string]int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var t time.Time
		var g, d sql.NullInt64
		err = rows.Scan(&t, &g, &d)
		if err != nil {
			return []map[string]int{}, err
		}
		gmap[t.Format("2006-01-02 15:04:05")] = int(g.Int64)
		dmap[t.Format("2006-01-02 15:04:05")] = int(d.Int64)
	}

	return []map[string]int{gmap, dmap}, nil
}

func (m *postgresDBRepo) GetSeriesDataFromVisitor(t time.Time, id int) ([]map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pmap := make(map[string]int)
	umap := make(map[string]int)
	stmt := `SELECT DATE_TRUNC('hour', timestamp) AS interval,
			 COALESCE(COUNT(anonymous_id), 0) ,COALESCE(COUNT(DISTINCT product_id),0)
			 FROM visitor
			 WHERE store = $1
			 AND timestamp >= $2
			 GROUP BY interval
			 ORDER BY interval;`
	rows, err := m.DB.QueryContext(ctx, stmt, id, t)
	if err != nil {
		return []map[string]int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var t time.Time
		var u, p sql.NullInt64
		err = rows.Scan(&t, &u, &p)
		if err != nil {
			return []map[string]int{}, err
		}
		umap[t.Format("2006-01-02 15:04:05")] = int(u.Int64)
		pmap[t.Format("2006-01-02 15:04:05")] = int(p.Int64)
	}

	return []map[string]int{umap, pmap}, nil
}

func (m *postgresDBRepo) GetTopProducts(id int) ([]int64, []int, []int, []int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	prods := []int64{}
	users := []int{}
	discounts := []int{}
	gmv := []int{}

	stmt1 := `SELECT COUNT(anonymous_id) as users, product_id
			 FROM visitor 
			 WHERE store = $1
			 GROUP BY product_id 
			 ORDER BY users desc limit 5`

	rows, err := m.DB.QueryContext(ctx, stmt1, id)
	if err != nil {
		return prods, users, discounts, gmv, err
	}
	defer rows.Close()
	for rows.Next() {
		var p, u sql.NullInt64

		err = rows.Scan(&u, &p)
		if !p.Valid || err != nil {
			return prods, users, discounts, gmv, err
		}
		prods = append(prods, p.Int64)
		users = append(users, int(u.Int64))
	}

	stmt2 := ``
	var rows2 *sql.Rows
	if len(prods) == 0 {
		m.App.InfoLog.Println("No products in the list")
		return prods, users, discounts, gmv, err
	} else if len(prods) < 5 && len(prods) >= 1 {
		/* There are less than 5 products, so just diplay the top 1 */
		stmt2 = `SELECT SUM(gmv), SUM(discount_amount)
			  	 FROM checkout
			  	 WHERE store = $1 and product_id = $2`
		rows2, err = m.DB.QueryContext(ctx, stmt2, id, prods[0])
		if err != nil {
			return prods, users, discounts, gmv, err
		}
		defer rows2.Close()
	} else {
		stmt2 = `SELECT product_id, SUM(gmv) AS total_gmv, SUM(discount_amount) AS total_discount
				 FROM (
				 	SELECT
				 		product_id,
						gmv,
						discount_amount,
						CASE
							WHEN product_id = $2 THEN 1
							WHEN product_id = $3 THEN 2
							WHEN product_id = $4 THEN 3
							WHEN product_id = $5 THEN 4
							WHEN product_id = $6 THEN 5
							ELSE 6  -- This is to handle other product_ids not in the IN clause
						END AS sort_order
					FROM checkout
					WHERE store = $1 and product_id IN ($2, $3, $4, $5, $6)
				 ) AS subquery
				 GROUP BY product_id, sort_order
				 ORDER BY sort_order`

		rows2, err = m.DB.QueryContext(ctx, stmt2, id,
			prods[0], prods[1], prods[2], prods[3], prods[4])
		if err != nil {
			return prods, users, discounts, gmv, err
		}
		defer rows2.Close()
	}

	for rows2.Next() {
		var p, g, d, s sql.NullInt64

		err = rows.Scan(&p, &g, &d, &s)
		if !g.Valid || err != nil {
			return prods, users, discounts, gmv, err
		}
		gmv = append(gmv, int(g.Int64))
		discounts = append(discounts, int(d.Int64))
	}

	return prods, users, discounts, gmv, nil
}

func (m *postgresDBRepo) GetAggOtfByDuration(t time.Time, id int) (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	otf := make(map[string]int)
	stmt := `SELECT DATE_TRUNC('hour', timestamp) AS interval,
			 COALESCE(COUNT(anonymous_id), 0)
			 FROM visitor
			 WHERE store = $1
			 AND timestamp >= $2
			 GROUP BY interval
			 ORDER BY interval`
	rows, err := m.DB.QueryContext(ctx, stmt, id, t)
	if err != nil {
		return map[string]int{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var t time.Time
		var u sql.NullInt64
		err = rows.Scan(&t, &u)
		if err != nil {
			return map[string]int{}, err
		}
		otf[t.Format("2006-01-02 15:04:05")] = int(u.Int64)
	}

	return otf, nil
}

func (m *postgresDBRepo) GetAllCampaigns(id int) ([]models.Camapign, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT
    		 DATE_TRUNC('day', visitor.timestamp) AS time_window,
    		 COALESCE(SUM(checkout.discount_amount),0) AS discount,
			 COALESCE(SUM(checkout.gmv),0) AS gmv,
			 COALESCE(COUNT(*),0) as users,
			 COUNT(DISTINCT visitor.product_id) AS products,
			 AVG(checkout.gmv) AS aov,
			 COUNT(CASE WHEN visitor.deal_shown = true THEN 1 ELSE NULL END) AS impressions,
			 COUNT(CASE WHEN visitor.code_copied = true THEN 1 ELSE NULL END) AS promo_copied,
			 COUNT(CASE WHEN checkout.gmv IS NOT NULL THEN 1 ELSE NULL END) AS conversions

			 FROM visitor LEFT OUTER JOIN checkout
			 ON visitor.discount_code = checkout.discount_code AND visitor.store = checkout.store
			 
			 WHERE visitor.timestamp >= $1 AND visitor.store = $2
			 
			 GROUP BY time_window
			 ORDER BY time_window DESC`

	j := []models.Camapign{}

	t, err := m.GetCampignEndTime(id)
	t.AddDate(-1, 0, 0)
	rows, err := m.DB.QueryContext(ctx, stmt, t, id)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Camapign
		err = rows.Scan(&c.StartTime, &c.DiscountValue, &c.GmvValue, &c.Users,
			&c.Products, &c.Aov, &c.Impressions, &c.PromoCopied, &c.Conversions)
		if err != nil {
			return j, err
		}
		j = append(j, c)
	}

	return j, nil
}

func (m *postgresDBRepo) UpdateDealListConfig(id int, md int8, pc string, bs int8, bc string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE store
			 SET max_discount_for_popup = $2, popup_color_code = $3, button_style = $4, button_color_code = $5
			 WHERE id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id, md, pc, bs, bc)
	if err != nil {
		m.App.ErrorLog.Println("DB insertion failed")
		return err
	}
	return nil
}

func (m *postgresDBRepo) GetStoreByID(id int) (models.Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT *  
			 FROM store 
			 WHERE id = $1`
	j := models.Store{}
	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		var crt, ctt string
		err = rows.Scan(&j.ID, &j.Name, &j.ApiToken, &j.RefreshToken, &j.Misc, &j.URL,
			&j.PopupColorCode, &j.ButtonColorCode, &j.DefaultDiscount,
			&j.DiscountCateogry, &j.MaxDiscountforPopup, &j.ButtonStyle,
			&crt, &ctt, &j.DealListActive, &j.Currency)
		if err != nil {
			return j, err
		}
	}
	return j, nil
}

// // START HERE

// /* Table "public.users"
// 	Column   |   Type
// -------------+-----------
//  id          | integer
//  firstname   | text
//  lastname    | text
//  email       | text
//  password    | text
//  accesslevel | integer
//  createdat   | timestamp without time zone
//  updatedat   | timestamp without time zone
//  storeid     | integer
// */

// func (m *postgresDBRepo) CreateUser(u models.User) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := `insert into users
// 				(firstname, lastname, email, password,
// 				accesslevel, createdat, updatedat, storeid)
// 			 values
// 			 	($1, $2, $3, $4, $5, $6, $7, $8)`

// 	_, err := m.DB.ExecContext(ctx, stmt, u.FirstName, u.LastName, u.Email,
// 		u.Password, u.AccessLevel, u.CreatedAt,
// 		u.UpdatedAt, u.Store.ID)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (m *postgresDBRepo) UpdateUser(u models.User) (models.User, error) {
// 	return models.User{}, nil
// }

// func (m *postgresDBRepo) GetUserByStore(storeid int) (models.User, error) {
// 	return models.User{}, nil
// }

// /* Table "public.price_rule"
//       Column       |  Type
// -------------------+--------
//  id                | integer
//  targettype        | text
//  targetselection   | text
//  valuetype         | text
//  value             | numeric
//  customerselection | text
//  allocationmethod  | text
//  startsat          | timestamp without time zone
// */

// func (m *postgresDBRepo) CreatePr(pr models.PriceRule) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := `insert into price_rule
// 				(prid, targettype, targetselection, valuetype,
// 				value, customerselection, allocationmethod, startsat)
// 			 values
// 			 	($1, $2, $3, $4, $5, $6, $7, $8)`

// 	_, err := m.DB.ExecContext(ctx, stmt, pr.ID, pr.TargetType, pr.TargetSelection,
// 		pr.ValueType, pr.Value, pr.CustomerSelection, pr.AllocationMethod, pr.StartsAt)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (m *postgresDBRepo) DeletePr(pr models.PriceRule) error {
// 	return nil
// }
// func (m *postgresDBRepo) GetPrById(id int64, storeid int) (models.PriceRule, error) {
// 	return models.PriceRule{}, nil
// }
// func (m *postgresDBRepo) ListPr(storeid int) ([]models.PriceRule, error) {
// 	return []models.PriceRule{}, nil
// }

// /* Table "public.discount_code"
// 	Column   |   Type
//  ------------+-----------
//  id          | integer
//  priceruleid | integer
//  code        | text
//  usagecount  | integer
//  createdat   | timestamp without time zone
//  updatedat   | timestamp without time zone
// */

// func (m *postgresDBRepo) CreateDiscountCode(d models.DiscountCode) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := `insert into discount_code
// 				(priceruleId, code, usagecount,
// 				createdat, updatedat)
// 			 values
// 			 	($1, $2, $3, $4, $5)`

// 	_, err := m.DB.ExecContext(ctx, stmt, d.PriceRuleID, d.Code,
// 		d.UsageCount, d.Timestamp, d.Timestamp)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (m *postgresDBRepo) DeleteDiscountCode(d models.DiscountCode) error {
// 	return nil
// }

// func (m *postgresDBRepo) GetDiscountsByPr(pr models.PriceRule) ([]models.DiscountCode, error) {
// 	return []models.DiscountCode{}, nil
// }

// /* Table "public.buyer"
//    Column    |  Type
// -------------+--------
//  anonymousid | integer
//  email       | text
//  storeid     | integer
//  productid   | integer
//  timestamp   | timestamp without time zone
//  gotdeal     | boolean
//  clickeddeal | boolean
//  cpid        | integer
//  misc        | text
// */

// func (m *postgresDBRepo) CreateBuyer(b models.Buyer) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := `insert into buyer
// 				(anonymousid, email, storeid, productid,
// 					timestamp, gotdeal, clickeddeal, cpid, misc)
// 			 values
// 			 	($1, $2, $3, $4, $5, $6, $7, $8, $9)`

// 	_, err := m.DB.ExecContext(ctx, stmt, b.AnonymousID, b.Email,
// 		b.Store.ID, b.ProductId, b.Timestamp, b.GotDeal, b.ClickedDeal,
// 		b.CPID, b.Misc)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (m *postgresDBRepo) GetBuyersByStore(storeid int) ([]models.Buyer, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := "select * from buyer where storeid = $1"
// 	j := []models.Buyer{}
// 	rows, err := m.DB.QueryContext(ctx, stmt, storeid)
// 	if err != nil {
// 		return j, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var b models.Buyer
// 		err = rows.Scan(&b.AnonymousID, &b.Email, &b.Store.ID, &b.ProductId,
// 			&b.Timestamp, &b.GotDeal, &b.ClickedDeal, &b.CPID, &b.Misc)
// 		if err != nil {
// 			return j, err
// 		}
// 		j = append(j, b)
// 	}
// 	return j, nil
// }

// func (m *postgresDBRepo) UpdateBuyer(b models.Buyer) (models.Buyer, error) {
// 	return models.Buyer{}, nil
// }

// func (m *postgresDBRepo) GetAggregateOtfByDuration(ts time.Time, typ string, id int) (map[string]int, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()
// 	otfMap := map[string]int{}

// 	stmt := `SELECT DATE_TRUNC('hour', timestamp) AS interval,
// 	COALESCE(COUNT(*), 0)
// 	FROM   buyer
// 	WHERE  storeid = $1
// 	AND	  timestamp >= $2
// 	GROUP BY interval
// 	ORDER BY interval;`
// 	rows, err := m.DB.QueryContext(ctx, stmt, id, ts)
// 	if err != nil {
// 		return otfMap, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var t time.Time
// 		var otf sql.NullInt64
// 		err = rows.Scan(&t, &otf)
// 		if err != nil {
// 			return otfMap, err
// 		}
// 		otfMap[t.Format("2006-01-02 15:04:05")] = int(otf.Int64)
// 	}

// 	return otfMap, nil
// }

// /* Table "public.store"
//     Column    |  Type
// --------------+--------
//  id           | integer
//  name         | text
//  apitoken     | text
//  refreshtoken | text
//  misc         | text
//  url          | text
// */

// func (m *postgresDBRepo) CreateStore(s models.Store) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := `insert into store
// 				(name, apitoken, refreshtoken,
// 				misc, url)
// 			 values
// 			 	($1, $2, $3, $4, $5)`

// 	_, err := m.DB.ExecContext(ctx, stmt, s.Name, s.ApiToken,
// 		s.RefreshToken, s.Misc, s.URL)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (m *postgresDBRepo) GetStoreByID(id int) (models.Store, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := "select * from store where id = $1"
// 	j := models.Store{}
// 	rows, err := m.DB.QueryContext(ctx, stmt, id)
// 	if err != nil {
// 		return j, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		err = rows.Scan(&j.ID, &j.Name, &j.ApiToken, &j.RefreshToken, &j.Misc, &j.URL)
// 		if err != nil {
// 			return j, err
// 		}
// 	}
// 	return j, nil
// }

// func (m *postgresDBRepo) UpdateStore(s models.Store) (models.Store, error) {
// 	return models.Store{}, nil
// }

// /* Table "public.campaign_product"
//     Column    |  Type
// --------------+--------
//  id           | integer
//  campaignid   | integer
//  productid    | integer
//  title        | text
//  storeid      | integer
//  deals        | integer
//  sold         | integer
//  dealdiscount | integer
//  emailsentto  | text[]
//  misc         | text
//  priceruleid  | integer
//  timestamp    | timestamp without time zone
//  price        | integer
// */

// func (m *postgresDBRepo) CreateCampaignProduct(cp models.CampaignProduct) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := `insert into campaign_product
// 				(campaignid, productid, title, storeid,
// 				deals, sold, dealdiscount, misc,
// 				timestamp, price)
// 			 values
// 			 	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

// 	_, err := m.DB.ExecContext(ctx, stmt, cp.CampaignID, cp.ProductID,
// 		cp.Title, cp.Store.ID, cp.Deals, cp.Sold, cp.DealDiscount,
// 		cp.Misc, cp.Timestamp, cp.Price)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (m *postgresDBRepo) GetCampaignProducts(c int64) ([]models.CampaignProduct, error) {
// 	// pull all campaign products with the given campaing ID from DB
// 	/* SELECT *
// 	FROM campaign_product
// 	WHERE CampaignID = 1;
// 	*/
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	stmt := `select * from campaign_product where
// 			campaignid = $1`

// 	rows, err := m.DB.QueryContext(ctx, stmt, c)
// 	if err != nil {
// 		return []models.CampaignProduct{}, err
// 	}
// 	defer rows.Close()

// 	var campaignProducts []models.CampaignProduct

// 	for rows.Next() {
// 		var cp models.CampaignProduct
// 		err := rows.Scan(&cp.ID, &cp.CampaignID, &cp.ProductID, &cp.Title,
// 			&cp.Store.ID, &cp.Deals, &cp.Sold, &cp.DealDiscount,
// 			&cp.EmailSentTo, &cp.Misc, &cp.PriceRuleID)
// 		if err != nil {
// 			return campaignProducts, err
// 		}
// 		campaignProducts = append(campaignProducts, cp)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return campaignProducts, err
// 	}

// 	return campaignProducts, nil
// }

// func (m *postgresDBRepo) UpdateCampaignProducts(c models.Campaign, dict map[string]interface{}) ([]models.CampaignProduct, error) {
// 	return []models.CampaignProduct{}, nil
// }

// func (m *postgresDBRepo) GetTopProductsByStore(s int) ([]models.CampaignProduct, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()
// 	/* SELECT * FROM products
// 	ORDER BY price ASC
// 	LIMIT 5;
// 	*/
// 	products := []models.CampaignProduct{}
// 	stmt := `select productid, deals, dealdiscount, price
// 	from campaign_product where storeid = $1 order by deals desc limit 5`
// 	rows, err := m.DB.QueryContext(ctx, stmt, s)
// 	if err != nil {
// 		return products, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var pid, d, dd, p sql.NullInt64

// 		err = rows.Scan(&pid, &d, &dd, &p)
// 		if !pid.Valid || err != nil {
// 			return products, err
// 		}
// 		var c models.CampaignProduct
// 		c.ProductID = pid.Int64
// 		c.Deals = int(d.Int64)
// 		c.DealDiscount = int(dd.Int64)
// 		c.Price = int(p.Int64)
// 		products = append(products, c)
// 	}

// 	return products, nil
// }
