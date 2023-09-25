package dbrepo

import (
	"context"
	"time"

	"github.com/malalwan/slaash/internal/models"
)

func (m *clickhouseDBRepo) AllUsers() {
}

func (m *clickhouseDBRepo) PullStreamByAnonymousID(id string) (models.VisitTable, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `select
				COUNT(event) FILTER (WHERE event = '$autocapture')
			from Clickstream
			where properties.$device_id = '$1'`
	j := models.VisitTable{}
	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return j, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&j.NumClicks)
		if err != nil {
			return j, err
		}
	}
	return j, nil
}
