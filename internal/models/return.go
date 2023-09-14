package models

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
type AggregateStats struct {
	Discount struct {
		Value int
	}
	GmvActiveSession struct {
		CurrencyType string
		Value        int
		GmvVertical  struct {
			Positive         bool
			ChangePercentage int
		}
	}
	ProductsActiveSession struct {
		Products         int
		ProductsVertical struct {
			Positive         bool
			ChangePercentage int
		}
	}
	ActiveUsers struct {
		ActiveUsersInSession int
		ActiveUsersVertical  struct {
			Positive         bool
			ChangePercentage int
		}
	}
	ActiveCampaignID int64
}

/* Json map for graphs + aggregate for overall stats */
type AggregateForGraphs struct {
	Gmv struct {
		Price       int
		GmvVertical struct {
			VerticalVal int
			GmvVertical bool
		}
	}
	Products struct {
		Price            int
		ProductsVertical struct {
			VerticalVal      int
			ProductsVertical bool
		}
	}
	Users struct {
		Price         int
		UsersVertical struct {
			VerticalVal   int
			UsersVertical string
		}
	}
	DiscountSpends struct {
		Price                  int
		DiscountSpendsVertical struct {
			VerticalVal            int
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
