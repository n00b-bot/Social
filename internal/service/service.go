package service

import (
	"database/sql"
	"sync"

	"github.com/hako/branca"
)

type Service struct {
	db                  *sql.DB
	codec               *branca.Branca
	timelineItemClients sync.Map
	commentClient       sync.Map
	notificationClient  sync.Map
}

func New(db *sql.DB, codec *branca.Branca) *Service {
	service := &Service{
		db:    db,
		codec: codec,
	}
	return service
}
