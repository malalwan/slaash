package dbrepo

import (
	"database/sql"

	"github.com/malalwan/slaash/internal/config"
	"github.com/malalwan/slaash/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type clickhouseDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

func NewClickhouseRepo(conn *sql.DB, a *config.AppConfig) repository.ClickhouseRepo {
	return &clickhouseDBRepo{
		App: a,
		DB:  conn,
	}
}
