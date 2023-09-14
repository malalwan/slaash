package models

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

type OtfResponse struct {
	Otf map[string]int
}
