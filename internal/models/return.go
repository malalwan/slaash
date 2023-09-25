package models

import "time"

/* This file contains structures used to
send data back to the dashboard as Json
All unique responses need to be defined
here for better visibility on the handlers
*/

/* Json map for top 5 products list */
type TopProducts struct {
	Products []struct {
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
}

/* Json map for active campaign stats */
type CampaignActivity struct {
	Discount struct {
		Value int
	}
	GmvActiveSession struct {
		CurrencyType string
		Value        int
		GmvVertical  struct {
			Positive         bool
			ChangePercentage float32
		}
	}
	ProductsActiveSession struct {
		Products         int
		ProductsVertical struct {
			Positive         bool
			ChangePercentage float32
		}
	}
	ActiveUsers struct {
		ActiveUsersInSession int
		ActiveUsersVertical  struct {
			Positive         bool
			ChangePercentage float32
		}
	}
	CampaignEndTime struct {
		Value  string
		Nextin int
	}
}

/* Json map for graphs + aggregate for overall stats */
type DealListActivity struct {
	Gmv struct {
		Price       int
		GmvVertical struct {
			VerticalVal float32
			GmvVertical bool
		}
	}
	Products struct {
		Price            int
		ProductsVertical struct {
			VerticalVal      float32
			ProductsVertical bool
		}
	}
	Users struct {
		Price         int
		UsersVertical struct {
			VerticalVal   float32
			UsersVertical bool
		}
	}
	DiscountSpends struct {
		Price                  int
		DiscountSpendsVertical struct {
			VerticalVal            float32
			DiscountSpendsVertical bool
		}
	}
	GmvData map[string]int

	ProductsData map[string]int

	UsersData map[string]int

	DiscountsData map[string]int
}

/* Json for OTF graph */
type OtfResponse struct {
	Otf map[string]int
}

/* VisitTable is the mapping for OTF algorithm and is used to cache that info in Postgres */
type VisitTable struct {
	AnonymousID      string
	FavClick         string
	MaxScrollDepth   int8
	ScrollVicinity   int8
	ScrollMap        map[int8]int16
	IndividualVisits int8
	LastAction       string
	LastActionTime   time.Time
	AddToCarts       int8
	CartItemsOmw     map[string]int8
	ImageClicks      int32
	StoreRoot        string
	Images           map[string]int16
	FavImage         string
	Referrer         string
	Pages            map[string]int16
	NumPages         int16
	NumClicks        int16
	AvgClickDistance float64
	ClickMap         map[string]int16
	LastClickTime    time.Time
	FavPage          string
}

type Camapign struct {
	StartTime             time.Time
	EndTime               time.Time
	DiscountValue         float32
	GmvValue              float32
	Users                 int
	Products              int
	Aov                   float32
	Impressions           int64
	PromoCopied           int64
	SuccessfulRedemptions int64
	Conversions           int64
}
