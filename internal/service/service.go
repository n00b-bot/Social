package service

import (
	"database/sql"

	"github.com/hako/branca"
)

type Service struct {
	db    *sql.DB
	codec *branca.Branca
}

func  New(db *sql.DB, codec *branca.Branca) *Service {
	service := &Service{
		db:    db,
		codec: codec,
	}
	return service
}
